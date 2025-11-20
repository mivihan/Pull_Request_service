package service

import (
	"context"

	"github.com/mivihan/Pull_Request_service/internal/domain"
	"github.com/mivihan/Pull_Request_service/internal/repository"
)

type UserService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetReviews(ctx context.Context, userID string) ([]*domain.PullRequest, error)
}

type userService struct {
	repos *repository.Repositories
}

func NewUserService(repos *repository.Repositories) UserService {
	return &userService{repos: repos}
}

func (s *userService) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	return s.repos.User.SetIsActive(ctx, userID, isActive)
}

func (s *userService) GetReviews(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	_, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.repos.PR.ListByReviewer(ctx, userID)
}
