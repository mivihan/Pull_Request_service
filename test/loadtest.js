import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  scenarios: {
    smoke: {
      executor: 'constant-vus',
      vus: 1,
      duration: '30s',
      tags: { test_type: 'smoke' },
      exec: 'smokeTest',
      startTime: '0s',
    },
    baseline: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 5 },
        { duration: '1m', target: 5 },
        { duration: '30s', target: 0 },
      ],
      tags: { test_type: 'baseline' },
      exec: 'baselineTest',
      startTime: '35s',
    },
    stress: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 10 },
        { duration: '1m', target: 20 },
        { duration: '30s', target: 30 },
        { duration: '30s', target: 0 },
      ],
      tags: { test_type: 'stress' },
      exec: 'stressTest',
      startTime: '3m',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<300'],
    http_req_failed: ['rate<0.01'],
    errors: ['rate<0.01'],
  },
};

function generateID(prefix) {
  return `${prefix}-${__VU}-${__ITER}-${Date.now()}`;
}

function createTeam(teamName, userCount = 3) {
  const members = [];
  for (let i = 0; i < userCount; i++) {
    members.push({
      user_id: generateID(`user`),
      username: `User${__VU}_${i}`,
      is_active: true,
    });
  }

  const payload = JSON.stringify({
    team_name: teamName,
    members: members,
  });

  const res = http.post(`${BASE_URL}/team/add`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  const success = check(res, {
    'team created': (r) => r.status === 201,
  });

  if (!success) {
    errorRate.add(1);
  }

  return { teamName, members, response: res };
}

function createPR(authorID) {
  const prID = generateID('pr');
  const payload = JSON.stringify({
    pull_request_id: prID,
    pull_request_name: `Feature ${prID}`,
    author_id: authorID,
  });

  const res = http.post(`${BASE_URL}/pullRequest/create`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  const success = check(res, {
    'PR created': (r) => r.status === 201,
  });

  if (!success) {
    errorRate.add(1);
  }

  return { prID, response: res };
}

function mergePR(prID) {
  const payload = JSON.stringify({
    pull_request_id: prID,
  });

  const res = http.post(`${BASE_URL}/pullRequest/merge`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  const success = check(res, {
    'PR merged': (r) => r.status === 200,
  });

  if (!success) {
    errorRate.add(1);
  }

  return res;
}

function getReviews(userID) {
  const res = http.get(`${BASE_URL}/users/getReview?user_id=${userID}`);

  const success = check(res, {
    'reviews retrieved': (r) => r.status === 200,
  });

  if (!success) {
    errorRate.add(1);
  }

  return res;
}

function getReviewerStats() {
  const res = http.get(`${BASE_URL}/stats/reviewers`);

  const success = check(res, {
    'stats retrieved': (r) => r.status === 200,
  });

  if (!success) {
    errorRate.add(1);
  }

  return res;
}

export function smokeTest() {
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health check ok': (r) => r.status === 200,
  });

  sleep(1);
}

export function baselineTest() {
  const team = createTeam(generateID('team'), 4);
  sleep(0.5);

  if (team.response.status === 201 && team.members.length > 0) {
    const pr = createPR(team.members[0].user_id);
    sleep(0.5);

    if (pr.response.status === 201) {
      mergePR(pr.prID);
      sleep(0.5);
    }

    if (team.members.length > 1) {
      getReviews(team.members[1].user_id);
      sleep(0.5);
    }
  }

  getReviewerStats();
  sleep(1);
}

export function stressTest() {
  const team = createTeam(generateID('team'), 5);

  if (team.response.status === 201 && team.members.length > 0) {
    const prs = [];
    for (let i = 0; i < 2; i++) {
      const pr = createPR(team.members[i % team.members.length].user_id);
      if (pr.response.status === 201) {
        prs.push(pr.prID);
      }
      sleep(0.1);
    }

    prs.forEach((prID, index) => {
      if (index % 2 === 0) {
        mergePR(prID);
        sleep(0.1);
      }
    });

    team.members.forEach((member) => {
      getReviews(member.user_id);
      sleep(0.1);
    });
  }

  getReviewerStats();
  sleep(0.5);
}