package domain

import (
	"fmt"
	"strings"
	"time"
)

type PullRequest struct {
	PullRequestID     string
	PullRequestName   string
	AuthorID          string
	Status            PRStatus
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          *time.Time
}

func (pr *PullRequest) Validate() error {
	if strings.TrimSpace(pr.PullRequestID) == "" {
		return fmt.Errorf("pull_request_id cannot be empty")
	}
	if strings.TrimSpace(pr.PullRequestName) == "" {
		return fmt.Errorf("pull_request_name cannot be empty")
	}
	if strings.TrimSpace(pr.AuthorID) == "" {
		return fmt.Errorf("author_id cannot be empty")
	}
	if !pr.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", pr.Status)
	}
	if len(pr.AssignedReviewers) > 2 {
		return fmt.Errorf("cannot have more than 2 reviewers")
	}
	return nil
}

func (pr *PullRequest) IsMerged() bool {
	return pr.Status == PRStatusMerged
}

func (pr *PullRequest) CanModifyReviewers() error {
	if pr.IsMerged() {
		return ErrPRMerged
	}
	return nil
}

func (pr *PullRequest) HasReviewer(userID string) bool {
	for _, reviewerID := range pr.AssignedReviewers {
		if reviewerID == userID {
			return true
		}
	}
	return false
}

func (pr *PullRequest) Merge() error {
	if pr.IsMerged() {
		return nil
	}

	pr.Status = PRStatusMerged
	now := time.Now()
	pr.MergedAt = &now

	return nil
}
