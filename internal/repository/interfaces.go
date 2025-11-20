package repository

import (
	"context"
	"time"

	"github.com/mivihan/Pull_Request_service/internal/domain"
)

type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error
	GetByName(ctx context.Context, teamName string) (*domain.Team, error)
	Exists(ctx context.Context, teamName string) (bool, error)
}

type UserRepository interface {
	Upsert(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	ListByTeam(ctx context.Context, teamName string) ([]*domain.User, error)
	ListActiveByTeamExcluding(ctx context.Context, teamName string, excludeUserIDs []string) ([]*domain.User, error)
}

type PRRepository interface {
	Create(ctx context.Context, pr *domain.PullRequest) error
	GetByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	Exists(ctx context.Context, prID string) (bool, error)
	UpdateStatus(ctx context.Context, prID string, status domain.PRStatus, mergedAt *time.Time) error
	AssignReviewers(ctx context.Context, prID string, userIDs []string) error
	ReplaceReviewer(ctx context.Context, prID string, oldUserID, newUserID string) error
	ListByReviewer(ctx context.Context, userID string) ([]*domain.PullRequest, error)
}

type Txer interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
