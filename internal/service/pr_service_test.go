package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mivihan/Pull_Request_service/internal/domain"
	"github.com/mivihan/Pull_Request_service/internal/repository"
)

type mockUserRepo struct {
	users map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[string]*domain.User),
	}
}

func (m *mockUserRepo) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	user, ok := m.users[userID]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepo) ListActiveByTeamExcluding(ctx context.Context, teamName string, excludeUserIDs []string) ([]*domain.User, error) {
	excludeMap := make(map[string]bool)
	for _, id := range excludeUserIDs {
		excludeMap[id] = true
	}

	var result []*domain.User
	for _, user := range m.users {
		if user.TeamName == teamName && user.IsActive && !excludeMap[user.UserID] {
			result = append(result, user)
		}
	}
	return result, nil
}

func (m *mockUserRepo) Upsert(ctx context.Context, user *domain.User) error {
	m.users[user.UserID] = user
	return nil
}

func (m *mockUserRepo) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserRepo) ListByTeam(ctx context.Context, teamName string) ([]*domain.User, error) {
	return nil, errors.New("not implemented")
}

type mockPRRepo struct {
	prs    map[string]*domain.PullRequest
	nextID int
}

func newMockPRRepo() *mockPRRepo {
	return &mockPRRepo{
		prs: make(map[string]*domain.PullRequest),
	}
}

func (m *mockPRRepo) Create(ctx context.Context, pr *domain.PullRequest) error {
	if _, exists := m.prs[pr.PullRequestID]; exists {
		return domain.ErrPRExists
	}
	m.prs[pr.PullRequestID] = pr
	return nil
}

func (m *mockPRRepo) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, ok := m.prs[prID]
	if !ok {
		return nil, domain.ErrPRNotFound
	}
	return pr, nil
}

func (m *mockPRRepo) Exists(ctx context.Context, prID string) (bool, error) {
	_, exists := m.prs[prID]
	return exists, nil
}

func (m *mockPRRepo) UpdateStatus(ctx context.Context, prID string, status domain.PRStatus, mergedAt *time.Time) error {
	pr, ok := m.prs[prID]
	if !ok {
		return domain.ErrPRNotFound
	}
	pr.Status = status
	pr.MergedAt = mergedAt
	return nil
}

func (m *mockPRRepo) AssignReviewers(ctx context.Context, prID string, userIDs []string) error {
	pr, ok := m.prs[prID]
	if !ok {
		return domain.ErrPRNotFound
	}
	pr.AssignedReviewers = userIDs
	return nil
}

func (m *mockPRRepo) ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	pr, ok := m.prs[prID]
	if !ok {
		return domain.ErrPRNotFound
	}

	found := false
	for i, id := range pr.AssignedReviewers {
		if id == oldUserID {
			pr.AssignedReviewers[i] = newUserID
			found = true
			break
		}
	}

	if !found {
		return domain.ErrNotAssigned
	}

	return nil
}

func (m *mockPRRepo) ListByReviewer(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	return nil, errors.New("not implemented")
}

type mockRepos struct {
	userRepo *mockUserRepo
	prRepo   *mockPRRepo
}

func (m *mockRepos) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func newMockRepos() *mockRepos {
	return &mockRepos{
		userRepo: newMockUserRepo(),
		prRepo:   newMockPRRepo(),
	}
}

func TestPRService_CreatePR_AutoAssign(t *testing.T) {
	mockRepos := newMockRepos()

	mockRepos.userRepo.users["u1"] = &domain.User{
		UserID:   "u1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}
	mockRepos.userRepo.users["u2"] = &domain.User{
		UserID:   "u2",
		Username: "Bob",
		TeamName: "backend",
		IsActive: true,
	}
	mockRepos.userRepo.users["u3"] = &domain.User{
		UserID:   "u3",
		Username: "Charlie",
		TeamName: "backend",
		IsActive: true,
	}

	repos := &repository.Repositories{}
	repos.User = mockRepos.userRepo
	repos.PR = mockRepos.prRepo

	service := NewPRService(repos)

	ctx := context.Background()
	pr, err := service.CreatePR(ctx, "pr-1", "Test PR", "u1")

	if err != nil {
		t.Fatalf("CreatePR failed: %v", err)
	}

	if pr.PullRequestID != "pr-1" {
		t.Errorf("expected PR ID pr-1, got %s", pr.PullRequestID)
	}

	if pr.Status != domain.PRStatusOpen {
		t.Errorf("expected status OPEN, got %s", pr.Status)
	}

	if len(pr.AssignedReviewers) > 2 {
		t.Errorf("expected max 2 reviewers, got %d", len(pr.AssignedReviewers))
	}

	for _, reviewerID := range pr.AssignedReviewers {
		if reviewerID == "u1" {
			t.Error("author should not be assigned as reviewer")
		}
	}
}

func TestPRService_CreatePR_NoActiveCandidates(t *testing.T) {
	mockRepos := newMockRepos()

	mockRepos.userRepo.users["u1"] = &domain.User{
		UserID:   "u1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	repos := &repository.Repositories{}
	repos.User = mockRepos.userRepo
	repos.PR = mockRepos.prRepo

	service := NewPRService(repos)

	ctx := context.Background()
	pr, err := service.CreatePR(ctx, "pr-1", "Test PR", "u1")

	if err != nil {
		t.Fatalf("CreatePR failed: %v", err)
	}

	if len(pr.AssignedReviewers) != 0 {
		t.Errorf("expected 0 reviewers, got %d", len(pr.AssignedReviewers))
	}
}

func TestPRService_MergePR_Idempotency(t *testing.T) {
	mockRepos := newMockRepos()

	mockRepos.prRepo.prs["pr-1"] = &domain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Test",
		AuthorID:          "u1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u2"},
		CreatedAt:         time.Now(),
	}

	repos := &repository.Repositories{}
	repos.PR = mockRepos.prRepo

	service := NewPRService(repos)

	ctx := context.Background()

	pr1, err := service.MergePR(ctx, "pr-1")
	if err != nil {
		t.Fatalf("first MergePR failed: %v", err)
	}

	if pr1.Status != domain.PRStatusMerged {
		t.Errorf("expected status MERGED, got %s", pr1.Status)
	}
	if pr1.MergedAt == nil {
		t.Error("expected merged_at to be set")
	}

	firstMergedAt := *pr1.MergedAt

	pr2, err := service.MergePR(ctx, "pr-1")
	if err != nil {
		t.Fatalf("second MergePR failed: %v", err)
	}

	if pr2.Status != domain.PRStatusMerged {
		t.Errorf("expected status to remain MERGED, got %s", pr2.Status)
	}

	if pr2.MergedAt == nil {
		t.Error("merged_at should still be set")
	}

	if !pr2.MergedAt.Equal(firstMergedAt) {
		t.Errorf("merged_at changed: %v != %v", *pr2.MergedAt, firstMergedAt)
	}
}

func TestPRService_ReassignReviewer_MergedPR(t *testing.T) {
	mockRepos := newMockRepos()

	now := time.Now()
	mockRepos.prRepo.prs["pr-1"] = &domain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Test",
		AuthorID:          "u1",
		Status:            domain.PRStatusMerged,
		AssignedReviewers: []string{"u2"},
		MergedAt:          &now,
	}

	mockRepos.userRepo.users["u2"] = &domain.User{
		UserID:   "u2",
		Username: "Bob",
		TeamName: "backend",
		IsActive: true,
	}

	repos := &repository.Repositories{}
	repos.User = mockRepos.userRepo
	repos.PR = mockRepos.prRepo

	service := NewPRService(repos)

	ctx := context.Background()
	_, _, err := service.ReassignReviewer(ctx, "pr-1", "u2")

	if err == nil {
		t.Fatal("expected error for merged PR, got nil")
	}

	if err != domain.ErrPRMerged {
		t.Errorf("expected ErrPRMerged, got %v", err)
	}
}

func TestPRService_ReassignReviewer_NotAssigned(t *testing.T) {
	mockRepos := newMockRepos()

	mockRepos.prRepo.prs["pr-1"] = &domain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Test",
		AuthorID:          "u1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u2"},
		CreatedAt:         time.Now(),
	}

	repos := &repository.Repositories{}
	repos.PR = mockRepos.prRepo

	service := NewPRService(repos)

	ctx := context.Background()

	_, _, err := service.ReassignReviewer(ctx, "pr-1", "u3")

	if err == nil {
		t.Fatal("expected error for not assigned user, got nil")
	}

	if err != domain.ErrNotAssigned {
		t.Errorf("expected ErrNotAssigned, got %v", err)
	}
}

func TestPRService_ReassignReviewer_NoCandidates(t *testing.T) {
	mockRepos := newMockRepos()

	mockRepos.prRepo.prs["pr-1"] = &domain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Test",
		AuthorID:          "u1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u2"},
		CreatedAt:         time.Now(),
	}

	mockRepos.userRepo.users["u1"] = &domain.User{
		UserID:   "u1",
		Username: "Alice",
		TeamName: "frontend",
		IsActive: true,
	}
	mockRepos.userRepo.users["u2"] = &domain.User{
		UserID:   "u2",
		Username: "Bob",
		TeamName: "backend",
		IsActive: true,
	}

	repos := &repository.Repositories{}
	repos.User = mockRepos.userRepo
	repos.PR = mockRepos.prRepo

	service := NewPRService(repos)

	ctx := context.Background()

	_, _, err := service.ReassignReviewer(ctx, "pr-1", "u2")

	if err == nil {
		t.Fatal("expected error for no candidates, got nil")
	}

	if err != domain.ErrNoCandidate {
		t.Errorf("expected ErrNoCandidate, got %v", err)
	}
}
