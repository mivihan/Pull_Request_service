package handler

import (
	"time"

	"github.com/mivihan/Pull_Request_service/internal/domain"
	"github.com/mivihan/Pull_Request_service/internal/service"
)

type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamDTO struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

type UserDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type PullRequestDTO struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         time.Time  `json:"createdAt"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type PullRequestShortDTO struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type CreateTeamRequest struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type TeamResponse struct {
	Team TeamDTO `json:"team"`
}

type UserResponse struct {
	User UserDTO `json:"user"`
}

type PRResponse struct {
	PR PullRequestDTO `json:"pr"`
}

type ReassignResponse struct {
	PR         PullRequestDTO `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}

type UserReviewsResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []PullRequestShortDTO `json:"pull_requests"`
}

func mapUserToDTO(u *domain.User) UserDTO {
	return UserDTO{
		UserID:   u.UserID,
		Username: u.Username,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
}

func mapTeamMemberToDTO(u *domain.User) TeamMemberDTO {
	return TeamMemberDTO{
		UserID:   u.UserID,
		Username: u.Username,
		IsActive: u.IsActive,
	}
}

func mapTeamWithMembersToDTO(t *service.TeamWithMembers) TeamDTO {
	members := make([]TeamMemberDTO, len(t.Members))
	for i, m := range t.Members {
		members[i] = mapTeamMemberToDTO(m)
	}
	return TeamDTO{
		TeamName: t.TeamName,
		Members:  members,
	}
}

func mapPRToDTO(pr *domain.PullRequest) PullRequestDTO {
	return PullRequestDTO{
		PullRequestID:     pr.PullRequestID,
		PullRequestName:   pr.PullRequestName,
		AuthorID:          pr.AuthorID,
		Status:            pr.Status.String(),
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

func mapPRToShortDTO(pr *domain.PullRequest) PullRequestShortDTO {
	return PullRequestShortDTO{
		PullRequestID:   pr.PullRequestID,
		PullRequestName: pr.PullRequestName,
		AuthorID:        pr.AuthorID,
		Status:          pr.Status.String(),
	}
}

type ReviewerStatDTO struct {
	UserID           string `json:"user_id"`
	AssignmentsCount int    `json:"assignments_count"`
}

type ReviewerStatsResponse struct {
	Reviewers []ReviewerStatDTO `json:"reviewers"`
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
	TeamName        string `json:"team_name"`
	DeactivatedCount int    `json:"deactivated_count"`
	AffectedPRCount  int    `json:"affected_pr_count"`
}