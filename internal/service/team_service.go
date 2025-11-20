package service

import (
	"context"
	"time"

	"github.com/mivihan/Pull_Request_service/internal/domain"
	"github.com/mivihan/Pull_Request_service/internal/repository"
)

type TeamMemberInput struct {
	UserID   string
	Username string
	IsActive bool
}

type TeamWithMembers struct {
	TeamName string
	Members  []*domain.User
}

type TeamService interface {
	CreateTeam(ctx context.Context, teamName string, members []TeamMemberInput) (*TeamWithMembers, error)
	GetTeam(ctx context.Context, teamName string) (*TeamWithMembers, error)
}

type teamService struct {
	repos *repository.Repositories
}

func NewTeamService(repos *repository.Repositories) TeamService {
	return &teamService{repos: repos}
}

func (s *teamService) CreateTeam(ctx context.Context, teamName string, members []TeamMemberInput) (*TeamWithMembers, error) {
	exists, err := s.repos.Team.Exists(ctx, teamName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrTeamExists
	}

	var resultMembers []*domain.User

	err = s.repos.WithTx(ctx, func(txCtx context.Context) error {
		team := &domain.Team{
			TeamName:  teamName,
			CreatedAt: time.Now(),
		}
		if err := s.repos.Team.Create(txCtx, team); err != nil {
			return err
		}

		for _, member := range members {
			user := &domain.User{
				UserID:    member.UserID,
				Username:  member.Username,
				TeamName:  teamName,
				IsActive:  member.IsActive,
				CreatedAt: time.Now(),
			}
			if err := s.repos.User.Upsert(txCtx, user); err != nil {
				return err
			}
			resultMembers = append(resultMembers, user)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &TeamWithMembers{
		TeamName: teamName,
		Members:  resultMembers,
	}, nil
}

func (s *teamService) GetTeam(ctx context.Context, teamName string) (*TeamWithMembers, error) {
	team, err := s.repos.Team.GetByName(ctx, teamName)
	if err != nil {
		return nil, err
	}
	members, err := s.repos.User.ListByTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}

	return &TeamWithMembers{
		TeamName: team.TeamName,
		Members:  members,
	}, nil
}
