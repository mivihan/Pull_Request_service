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

type DeactivationResult struct {
	TeamName         string
	DeactivatedCount int
	AffectedPRCount  int
}

type TeamService interface {
	CreateTeam(ctx context.Context, teamName string, members []TeamMemberInput) (*TeamWithMembers, error)
	GetTeam(ctx context.Context, teamName string) (*TeamWithMembers, error)
	DeactivateTeamUsers(ctx context.Context, teamName string, userIDs []string) (*DeactivationResult, error)
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

func (s *teamService) DeactivateTeamUsers(ctx context.Context, teamName string, userIDs []string) (*DeactivationResult, error) {
	if len(userIDs) == 0 {
		return &DeactivationResult{
			TeamName:         teamName,
			DeactivatedCount: 0,
			AffectedPRCount:  0,
		}, nil
	}
	_, err := s.repos.Team.GetByName(ctx, teamName)
	if err != nil {
		return nil, err
	}

	var deactivatedCount int
	var affectedPRCount int
	err = s.repos.WithTx(ctx, func(txCtx context.Context) error {
		count, err := s.repos.User.DeactivateUsers(txCtx, teamName, userIDs)
		if err != nil {
			return err
		}
		deactivatedCount = count

		if deactivatedCount == 0 {
			return nil
		}
		affectedPRs, err := s.repos.PR.GetOpenPRsByReviewers(txCtx, userIDs)
		if err != nil {
			return err
		}

		affectedPRCount = len(affectedPRs)
		deactivatedSet := make(map[string]bool)
		for _, uid := range userIDs {
			deactivatedSet[uid] = true
		}

		for _, pr := range affectedPRs {
			newReviewers := make([]string, 0, len(pr.AssignedReviewers))

			for _, reviewerID := range pr.AssignedReviewers {
				if deactivatedSet[reviewerID] {
					excludeIDs := append([]string{pr.AuthorID}, pr.AssignedReviewers...)
					
					candidates, err := s.repos.User.ListActiveByTeamExcluding(txCtx, teamName, excludeIDs)
					if err != nil {
						return err
					}

					if len(candidates) > 0 {
						replacement := candidates[0].UserID
						newReviewers = append(newReviewers, replacement)
						pr.AssignedReviewers = append(pr.AssignedReviewers, replacement)
					}
				} else {
					newReviewers = append(newReviewers, reviewerID)
				}
			}

			if err := s.repos.PR.AssignReviewers(txCtx, pr.PullRequestID, newReviewers); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &DeactivationResult{
		TeamName:         teamName,
		DeactivatedCount: deactivatedCount,
		AffectedPRCount:  affectedPRCount,
	}, nil
}