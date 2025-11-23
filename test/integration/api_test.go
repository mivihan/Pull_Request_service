package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const baseURL = "http://localhost:8080"

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type CreateTeamRequest struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type TeamResponse struct {
	Team struct {
		TeamName string       `json:"team_name"`
		Members  []TeamMember `json:"members"`
	} `json:"team"`
}

type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type PRResponse struct {
	PR struct {
		PullRequestID     string     `json:"pull_request_id"`
		PullRequestName   string     `json:"pull_request_name"`
		AuthorID          string     `json:"author_id"`
		Status            string     `json:"status"`
		AssignedReviewers []string   `json:"assigned_reviewers"`
		CreatedAt         time.Time  `json:"createdAt"`
		MergedAt          *time.Time `json:"mergedAt,omitempty"`
	} `json:"pr"`
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type UserReviewsResponse struct {
	UserID       string `json:"user_id"`
	PullRequests []struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
		Status          string `json:"status"`
	} `json:"pull_requests"`
}

type ReviewerStatsResponse struct {
	Reviewers []struct {
		UserID           string `json:"user_id"`
		AssignmentsCount int    `json:"assignments_count"`
	} `json:"reviewers"`
}

type PRStatsResponse struct {
	Open   int `json:"open"`
	Merged int `json:"merged"`
}

type DeactivateUsersRequest struct {
	TeamName string   `json:"team_name"`
	UserIDs  []string `json:"user_ids"`
}

type DeactivateUsersResponse struct {
	TeamName         string `json:"team_name"`
	DeactivatedCount int    `json:"deactivated_count"`
	AffectedPRCount  int    `json:"affected_pr_count"`
}

type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type UserResponse struct {
	User struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		TeamName string `json:"team_name"`
		IsActive bool   `json:"is_active"`
	} `json:"user"`
}

type TeamGetResponse struct {
    TeamName string       `json:"team_name"`
    Members  []TeamMember `json:"members"`
}

func makeRequest(t *testing.T, method, path string, body interface{}) (*http.Response, []byte) {
	t.Helper()

	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}
	}

	req, err := http.NewRequest(method, baseURL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody := new(bytes.Buffer)
	if _, err := respBody.ReadFrom(resp.Body); err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	return resp, respBody.Bytes()
}

func generateID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func TestHealthCheck(t *testing.T) {
	resp, body := makeRequest(t, "GET", "/health", nil)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if string(body) != "OK" {
		t.Errorf("expected 'OK', got '%s'", string(body))
	}
}

func TestCreateTeamAndPR(t *testing.T) {
	teamName := generateID("team")

	createTeamReq := CreateTeamRequest{
		TeamName: teamName,
		Members: []TeamMember{
			{UserID: generateID("user"), Username: "Alice", IsActive: true},
			{UserID: generateID("user"), Username: "Bob", IsActive: true},
			{UserID: generateID("user"), Username: "Charlie", IsActive: true},
		},
	}

	resp, body := makeRequest(t, "POST", "/team/add", createTeamReq)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("team creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var teamResp TeamResponse
	if err := json.Unmarshal(body, &teamResp); err != nil {
		t.Fatalf("failed to parse team response: %v", err)
	}

	if teamResp.Team.TeamName != teamName {
		t.Errorf("team name mismatch: expected %s, got %s", teamName, teamResp.Team.TeamName)
	}

	if len(teamResp.Team.Members) != 3 {
		t.Errorf("expected 3 members, got %d", len(teamResp.Team.Members))
	}

	authorID := teamResp.Team.Members[0].UserID
	prID := generateID("pr")

	createPRReq := CreatePRRequest{
		PullRequestID:   prID,
		PullRequestName: "Feature implementation",
		AuthorID:        authorID,
	}

	resp, body = makeRequest(t, "POST", "/pullRequest/create", createPRReq)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("PR creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var prResp PRResponse
	if err := json.Unmarshal(body, &prResp); err != nil {
		t.Fatalf("failed to parse PR response: %v", err)
	}

	if prResp.PR.Status != "OPEN" {
		t.Errorf("expected status OPEN, got %s", prResp.PR.Status)
	}

	if len(prResp.PR.AssignedReviewers) > 2 {
		t.Errorf("too many reviewers: expected max 2, got %d", len(prResp.PR.AssignedReviewers))
	}

	for _, reviewerID := range prResp.PR.AssignedReviewers {
		if reviewerID == authorID {
			t.Error("author should not be assigned as reviewer")
		}
	}

	memberIDs := make(map[string]bool)
	for _, member := range teamResp.Team.Members {
		memberIDs[member.UserID] = true
	}

	for _, reviewerID := range prResp.PR.AssignedReviewers {
		if !memberIDs[reviewerID] {
			t.Errorf("reviewer %s is not a team member", reviewerID)
		}
	}
}

func TestMergePRIdempotency(t *testing.T) {
	teamName := generateID("team")

	createTeamReq := CreateTeamRequest{
		TeamName: teamName,
		Members: []TeamMember{
			{UserID: generateID("user"), Username: "Alice", IsActive: true},
			{UserID: generateID("user"), Username: "Bob", IsActive: true},
		},
	}

	resp, body := makeRequest(t, "POST", "/team/add", createTeamReq)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("setup failed: %s", string(body))
	}

	var teamResp TeamResponse
	json.Unmarshal(body, &teamResp)

	prID := generateID("pr")
	createPRReq := CreatePRRequest{
		PullRequestID:   prID,
		PullRequestName: "Test merge",
		AuthorID:        teamResp.Team.Members[0].UserID,
	}

	resp, body = makeRequest(t, "POST", "/pullRequest/create", createPRReq)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("setup failed: %s", string(body))
	}

	mergeReq := MergePRRequest{PullRequestID: prID}
	resp, body = makeRequest(t, "POST", "/pullRequest/merge", mergeReq)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("first merge failed with status %d: %s", resp.StatusCode, string(body))
	}

	var prResp1 PRResponse
	json.Unmarshal(body, &prResp1)

	if prResp1.PR.Status != "MERGED" {
		t.Errorf("expected status MERGED, got %s", prResp1.PR.Status)
	}

	if prResp1.PR.MergedAt == nil {
		t.Fatal("merged_at should be set after merge")
	}

	firstMergedAt := *prResp1.PR.MergedAt

	time.Sleep(100 * time.Millisecond)

	resp, body = makeRequest(t, "POST", "/pullRequest/merge", mergeReq)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("second merge failed with status %d: %s", resp.StatusCode, string(body))
	}

	var prResp2 PRResponse
	json.Unmarshal(body, &prResp2)

	if prResp2.PR.Status != "MERGED" {
		t.Errorf("status should remain MERGED, got %s", prResp2.PR.Status)
	}

	if prResp2.PR.MergedAt == nil {
		t.Fatal("merged_at should still be set")
	}

	first := firstMergedAt.Truncate(time.Microsecond)
	second := prResp2.PR.MergedAt.Truncate(time.Microsecond)
	
	if !second.Equal(first) {
		t.Errorf("merged_at changed on second merge: %v -> %v", firstMergedAt, *prResp2.PR.MergedAt)
	}
}

func TestUserReviewsAndStats(t *testing.T) {
	teamName := generateID("team")

	createTeamReq := CreateTeamRequest{
		TeamName: teamName,
		Members: []TeamMember{
			{UserID: generateID("user"), Username: "Alice", IsActive: true},
			{UserID: generateID("user"), Username: "Bob", IsActive: true},
			{UserID: generateID("user"), Username: "Charlie", IsActive: true},
		},
	}

	resp, body := makeRequest(t, "POST", "/team/add", createTeamReq)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("setup failed: %s", string(body))
	}

	var teamResp TeamResponse
	json.Unmarshal(body, &teamResp)

	prID := generateID("pr")
	createPRReq := CreatePRRequest{
		PullRequestID:   prID,
		PullRequestName: "Feature",
		AuthorID:        teamResp.Team.Members[0].UserID,
	}

	resp, body = makeRequest(t, "POST", "/pullRequest/create", createPRReq)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("setup failed: %s", string(body))
	}

	var prResp PRResponse
	json.Unmarshal(body, &prResp)

	if len(prResp.PR.AssignedReviewers) == 0 {
		t.Skip("no reviewers assigned, skipping review check")
	}

	reviewerID := prResp.PR.AssignedReviewers[0]

	resp, body = makeRequest(t, "GET", fmt.Sprintf("/users/getReview?user_id=%s", reviewerID), nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("getReview failed with status %d: %s", resp.StatusCode, string(body))
	}

	var reviewsResp UserReviewsResponse
	if err := json.Unmarshal(body, &reviewsResp); err != nil {
		t.Fatalf("failed to parse reviews response: %v", err)
	}

	if reviewsResp.UserID != reviewerID {
		t.Errorf("user_id mismatch: expected %s, got %s", reviewerID, reviewsResp.UserID)
	}

	found := false
	for _, pr := range reviewsResp.PullRequests {
		if pr.PullRequestID == prID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("created PR %s not found in reviewer's list", prID)
	}

	resp, body = makeRequest(t, "GET", "/stats/reviewers", nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("stats/reviewers failed with status %d: %s", resp.StatusCode, string(body))
	}

	var statsResp ReviewerStatsResponse
	if err := json.Unmarshal(body, &statsResp); err != nil {
		t.Fatalf("failed to parse stats response: %v", err)
	}

	resp, body = makeRequest(t, "GET", "/stats/pullRequests", nil)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("stats/pullRequests failed with status %d: %s", resp.StatusCode, string(body))
	}

	var prStatsResp PRStatsResponse
	if err := json.Unmarshal(body, &prStatsResp); err != nil {
		t.Fatalf("failed to parse PR stats response: %v", err)
	}

	if prStatsResp.Open < 0 || prStatsResp.Merged < 0 {
		t.Errorf("stats should not be negative: open=%d, merged=%d", prStatsResp.Open, prStatsResp.Merged)
	}
}

func TestBulkDeactivateUsers(t *testing.T) {
    teamName := generateID("team")

    user1ID := generateID("user")
    user2ID := generateID("user")
    user3ID := generateID("user")
    user4ID := generateID("user")

    createTeamReq := CreateTeamRequest{
        TeamName: teamName,
        Members: []TeamMember{
            {UserID: user1ID, Username: "Alice", IsActive: true},
            {UserID: user2ID, Username: "Bob", IsActive: true},
            {UserID: user3ID, Username: "Charlie", IsActive: true},
            {UserID: user4ID, Username: "Dave", IsActive: true},
        },
    }

    resp, body := makeRequest(t, "POST", "/team/add", createTeamReq)
    if resp.StatusCode != http.StatusCreated {
        t.Fatalf("setup failed (team): %s", string(body))
    }

    prID := generateID("pr")
    createPRReq := CreatePRRequest{
        PullRequestID:   prID,
        PullRequestName: "Feature",
        AuthorID:        user1ID,
    }

    resp, body = makeRequest(t, "POST", "/pullRequest/create", createPRReq)
    if resp.StatusCode != http.StatusCreated {
        t.Fatalf("setup failed (PR): %s", string(body))
    }

    var prResp PRResponse
    if err := json.Unmarshal(body, &prResp); err != nil {
        t.Fatalf("failed to unmarshal PR response: %v", err)
    }
    initialReviewers := append([]string(nil), prResp.PR.AssignedReviewers...)

    deactivateReq := DeactivateUsersRequest{
        TeamName: teamName,
        UserIDs:  []string{user2ID, user3ID},
    }

    resp, body = makeRequest(t, "POST", "/team/deactivateUsers", deactivateReq)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("deactivation failed with status %d: %s", resp.StatusCode, string(body))
    }

    var deactivateResp DeactivateUsersResponse
    if err := json.Unmarshal(body, &deactivateResp); err != nil {
        t.Fatalf("failed to parse deactivation response: %v", err)
    }

    if deactivateResp.DeactivatedCount != 2 {
        t.Errorf("expected 2 users deactivated, got %d", deactivateResp.DeactivatedCount)
    }

    resp, body = makeRequest(t, "GET", "/team/get?team_name="+teamName, nil)
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("failed to get team after deactivation: %s", string(body))
    }

    var teamGetResp TeamGetResponse
	if err := json.Unmarshal(body, &teamGetResp); err != nil {
		t.Fatalf("failed to unmarshal team response: %v", err)
	}

	inactive := map[string]bool{user2ID: false, user3ID: false}
	for _, m := range teamGetResp.Members {
		if m.UserID == user2ID || m.UserID == user3ID {
			inactive[m.UserID] = !m.IsActive
		}
	}

    for id, ok := range inactive {
        if !ok {
            t.Errorf("expected user %s to be inactive after deactivation", id)
        }
    }

    resp, body = makeRequest(t, "POST", "/pullRequest/create", CreatePRRequest{
        PullRequestID:   generateID("pr"),
        PullRequestName: "Another feature",
        AuthorID:        user1ID,
    })
    if resp.StatusCode != http.StatusCreated {
        t.Fatalf("failed to create PR after deactivation: %s", string(body))
    }

    var newPRResp PRResponse
    if err := json.Unmarshal(body, &newPRResp); err != nil {
        t.Fatalf("failed to unmarshal new PR response: %v", err)
    }

    for _, reviewerID := range newPRResp.PR.AssignedReviewers {
        if reviewerID == user2ID || reviewerID == user3ID {
            t.Errorf("deactivated user %s should not be assigned as reviewer", reviewerID)
        }
    }

    t.Logf("Deactivation successful: %d users deactivated, %d PRs affected",
        deactivateResp.DeactivatedCount, deactivateResp.AffectedPRCount)
    t.Logf("Initial reviewers: %v", initialReviewers)
}