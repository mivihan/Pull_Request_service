package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/mivihan/Pull_Request_service/internal/domain"
	"github.com/mivihan/Pull_Request_service/internal/repository"
)

type PRService interface {
	CreatePR(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (*domain.PullRequest, string, error)
	GetReviewerStats(ctx context.Context) (map[string]int, error)
	GetPRStats(ctx context.Context) (map[string]int, error)
}

type prService struct {
	repos *repository.Repositories
	rand  *rand.Rand
}

func NewPRService(repos *repository.Repositories) PRService {
	return &prService{
		repos: repos,
		rand:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *prService) CreatePR(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {
	exists, err := s.repos.PR.Exists(ctx, prID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrPRExists
	}

	author, err := s.repos.User.GetByID(ctx, authorID)
	if err != nil {
		return nil, err
	}

	candidates, err := s.repos.User.ListActiveByTeamExcluding(ctx, author.TeamName, []string{authorID})
	if err != nil {
		return nil, err
	}

	reviewers := s.selectRandomReviewers(candidates, 2)
	reviewerIDs := extractUserIDs(reviewers)

	pr := &domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: reviewerIDs,
		CreatedAt:         time.Now(),
	}

	err = s.repos.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repos.PR.Create(txCtx, pr); err != nil {
			return err
		}
		if len(reviewerIDs) > 0 {
			if err := s.repos.PR.AssignReviewers(txCtx, prID, reviewerIDs); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *prService) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.repos.PR.GetByID(ctx, prID)
	if err != nil {
		return nil, err
	}

	if pr.IsMerged() {
		return pr, nil
	}

	pr.Merge()

	err = s.repos.PR.UpdateStatus(ctx, prID, pr.Status, pr.MergedAt)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *prService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*domain.PullRequest, string, error) {
	pr, err := s.repos.PR.GetByID(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	if err := pr.CanModifyReviewers(); err != nil {
		return nil, "", err
	}

	if !pr.HasReviewer(oldUserID) {
		return nil, "", domain.ErrNotAssigned
	}

	oldReviewer, err := s.repos.User.GetByID(ctx, oldUserID)
	if err != nil {
		return nil, "", err
	}

	excludeIDs := append(pr.AssignedReviewers, pr.AuthorID)
	candidates, err := s.repos.User.ListActiveByTeamExcluding(ctx, oldReviewer.TeamName, excludeIDs)
	if err != nil {
		return nil, "", err
	}

	if len(candidates) == 0 {
		return nil, "", domain.ErrNoCandidate
	}

	newReviewer := s.selectRandomReviewers(candidates, 1)[0]

	err = s.repos.WithTx(ctx, func(txCtx context.Context) error {
		return s.repos.PR.ReplaceReviewer(txCtx, prID, oldUserID, newReviewer.UserID)
	})
	if err != nil {
		return nil, "", err
	}

	for i, id := range pr.AssignedReviewers {
		if id == oldUserID {
			pr.AssignedReviewers[i] = newReviewer.UserID
			break
		}
	}

	return pr, newReviewer.UserID, nil
}

func (s *prService) selectRandomReviewers(users []*domain.User, maxCount int) []*domain.User {
	if len(users) == 0 {
		return nil
	}

	count := min(maxCount, len(users))

	shuffled := make([]*domain.User, len(users))
	copy(shuffled, users)
	s.rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}

func extractUserIDs(users []*domain.User) []string {
	ids := make([]string, len(users))
	for i, user := range users {
		ids[i] = user.UserID
	}
	return ids
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *prService) GetReviewerStats(ctx context.Context) (map[string]int, error) {
	return s.repos.PR.GetReviewerStats(ctx)
}

func (s *prService) GetPRStats(ctx context.Context) (map[string]int, error) {
	return s.repos.PR.GetPRStats(ctx)
}