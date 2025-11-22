package domain

import (
	"testing"
	"time"
)

func TestPullRequest_Merge_Idempotency(t *testing.T) {
	pr := &PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Test PR",
		AuthorID:          "u1",
		Status:            PRStatusOpen,
		AssignedReviewers: []string{"u2", "u3"},
		CreatedAt:         time.Now(),
	}

	pr.Merge()

	if pr.Status != PRStatusMerged {
		t.Errorf("expected status MERGED, got %s", pr.Status)
	}
	if pr.MergedAt == nil {
		t.Fatal("expected merged_at to be set")
	}

	firstMergedAt := *pr.MergedAt

	time.Sleep(10 * time.Millisecond)

	pr.Merge()

	if pr.Status != PRStatusMerged {
		t.Errorf("expected status to remain MERGED, got %s", pr.Status)
	}
	if pr.MergedAt == nil {
		t.Fatal("merged_at should still be set")
	}
	if !pr.MergedAt.Equal(firstMergedAt) {
		t.Errorf("merged_at changed on second merge: %v != %v", *pr.MergedAt, firstMergedAt)
	}
}

func TestPullRequest_CanModifyReviewers(t *testing.T) {
	tests := []struct {
		name    string
		status  PRStatus
		wantErr bool
		errType *DomainError
	}{
		{
			name:    "open PR can be modified",
			status:  PRStatusOpen,
			wantErr: false,
		},
		{
			name:    "merged PR cannot be modified",
			status:  PRStatusMerged,
			wantErr: true,
			errType: ErrPRMerged,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PullRequest{
				PullRequestID: "pr-test",
				Status:        tt.status,
			}

			err := pr.CanModifyReviewers()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("expected error %v, got %v", tt.errType, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestPullRequest_HasReviewer(t *testing.T) {
	pr := &PullRequest{
		PullRequestID:     "pr-1",
		AssignedReviewers: []string{"u2", "u3"},
	}

	tests := []struct {
		name     string
		userID   string
		expected bool
	}{
		{
			name:     "reviewer exists - first",
			userID:   "u2",
			expected: true,
		},
		{
			name:     "reviewer exists - second",
			userID:   "u3",
			expected: true,
		},
		{
			name:     "reviewer does not exist",
			userID:   "u1",
			expected: false,
		},
		{
			name:     "empty user id",
			userID:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pr.HasReviewer(tt.userID)
			if result != tt.expected {
				t.Errorf("HasReviewer(%s) = %v, want %v", tt.userID, result, tt.expected)
			}
		})
	}
}

func TestPullRequest_IsMerged(t *testing.T) {
	tests := []struct {
		name     string
		status   PRStatus
		expected bool
	}{
		{
			name:     "open PR is not merged",
			status:   PRStatusOpen,
			expected: false,
		},
		{
			name:     "merged PR is merged",
			status:   PRStatusMerged,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PullRequest{
				PullRequestID: "pr-test",
				Status:        tt.status,
			}

			result := pr.IsMerged()
			if result != tt.expected {
				t.Errorf("IsMerged() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPullRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		pr      *PullRequest
		wantErr bool
	}{
		{
			name: "valid PR",
			pr: &PullRequest{
				PullRequestID:     "pr-1",
				PullRequestName:   "Test PR",
				AuthorID:          "u1",
				Status:            PRStatusOpen,
				AssignedReviewers: []string{"u2"},
			},
			wantErr: false,
		},
		{
			name: "empty pull_request_id",
			pr: &PullRequest{
				PullRequestID:   "",
				PullRequestName: "Test",
				AuthorID:        "u1",
				Status:          PRStatusOpen,
			},
			wantErr: true,
		},
		{
			name: "empty pull_request_name",
			pr: &PullRequest{
				PullRequestID:   "pr-1",
				PullRequestName: "",
				AuthorID:        "u1",
				Status:          PRStatusOpen,
			},
			wantErr: true,
		},
		{
			name: "empty author_id",
			pr: &PullRequest{
				PullRequestID:   "pr-1",
				PullRequestName: "Test",
				AuthorID:        "",
				Status:          PRStatusOpen,
			},
			wantErr: true,
		},
		{
			name: "too many reviewers",
			pr: &PullRequest{
				PullRequestID:     "pr-1",
				PullRequestName:   "Test",
				AuthorID:          "u1",
				Status:            PRStatusOpen,
				AssignedReviewers: []string{"u2", "u3", "u4"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pr.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
