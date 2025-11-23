package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mivihan/Pull_Request_service/internal/domain"
)

type PostgresPRRepository struct {
	pool *pgxpool.Pool
}

func NewPRRepository(pool *pgxpool.Pool) PRRepository {
	return &PostgresPRRepository{pool: pool}
}

func (r *PostgresPRRepository) Create(ctx context.Context, pr *domain.PullRequest) error {
	if err := pr.Validate(); err != nil {
		return fmt.Errorf("invalid pull request: %w", err)
	}

	q := getQuerier(ctx, r.pool)

	query := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at, merged_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := q.Exec(ctx, query,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.AuthorID,
		pr.Status,
		pr.CreatedAt,
		pr.MergedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrPRExists
		}
		return fmt.Errorf("insert pull request: %w", err)
	}

	return nil
}

func (r *PostgresPRRepository) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	q := getQuerier(ctx, r.pool)

	prQuery := `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`

	var pr domain.PullRequest
	err := q.QueryRow(ctx, prQuery, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPRNotFound
		}
		return nil, fmt.Errorf("query pull request: %w", err)
	}

	reviewersQuery := `
		SELECT user_id
		FROM pr_reviewers
		WHERE pr_id = $1
		ORDER BY assigned_at
	`

	rows, err := q.Query(ctx, reviewersQuery, prID)
	if err != nil {
		return nil, fmt.Errorf("query reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan reviewer: %w", err)
		}
		reviewers = append(reviewers, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviewers: %w", err)
	}

	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *PostgresPRRepository) Exists(ctx context.Context, prID string) (bool, error) {
	q := getQuerier(ctx, r.pool)

	query := `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`

	var exists bool
	err := q.QueryRow(ctx, query, prID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check PR existence: %w", err)
	}

	return exists, nil
}

func (r *PostgresPRRepository) UpdateStatus(ctx context.Context, prID string, status domain.PRStatus, mergedAt *time.Time) error {
	q := getQuerier(ctx, r.pool)

	query := `
		UPDATE pull_requests
		SET status = $2, merged_at = $3
		WHERE pull_request_id = $1
	`

	result, err := q.Exec(ctx, query, prID, status, mergedAt)
	if err != nil {
		return fmt.Errorf("update PR status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrPRNotFound
	}

	return nil
}

func (r *PostgresPRRepository) AssignReviewers(ctx context.Context, prID string, userIDs []string) error {
	q := getQuerier(ctx, r.pool)

	deleteQuery := `DELETE FROM pr_reviewers WHERE pr_id = $1`
	_, err := q.Exec(ctx, deleteQuery, prID)
	if err != nil {
		return fmt.Errorf("delete existing reviewers: %w", err)
	}

	if len(userIDs) == 0 {
		return nil
	}

	insertQuery := `
		INSERT INTO pr_reviewers (pr_id, user_id, assigned_at)
		VALUES ($1, $2, $3)
	`

	now := time.Now()
	for _, userID := range userIDs {
		_, err := q.Exec(ctx, insertQuery, prID, userID, now)
		if err != nil {
			return fmt.Errorf("insert reviewer %s: %w", userID, err)
		}
	}

	return nil
}

func (r *PostgresPRRepository) ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	q := getQuerier(ctx, r.pool)

	query := `
		UPDATE pr_reviewers
		SET user_id = $3, assigned_at = $4
		WHERE pr_id = $1 AND user_id = $2
	`

	result, err := q.Exec(ctx, query, prID, oldUserID, newUserID, time.Now())
	if err != nil {
		return fmt.Errorf("replace reviewer: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotAssigned
	}

	return nil
}

func (r *PostgresPRRepository) ListByReviewer(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	q := getQuerier(ctx, r.pool)

	query := `
		SELECT DISTINCT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
		FROM pull_requests pr
		INNER JOIN pr_reviewers rev ON pr.pull_request_id = rev.pr_id
		WHERE rev.user_id = $1
		ORDER BY pr.created_at DESC
	`

	rows, err := q.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query PRs by reviewer: %w", err)
	}
	defer rows.Close()

	var prs []*domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		if err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&pr.Status,
			&pr.CreatedAt,
			&pr.MergedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pull request: %w", err)
		}
		prs = append(prs, &pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate PRs: %w", err)
	}

	return prs, nil
}

// GetReviewerStats returns how many times each user was assigned as reviewer
func (r *PostgresPRRepository) GetReviewerStats(ctx context.Context) (map[string]int, error) {
	q := getQuerier(ctx, r.pool)

	query := `
		SELECT user_id, COUNT(*) as assignments_count
		FROM pr_reviewers
		GROUP BY user_id
		ORDER BY assignments_count DESC
	`

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query reviewer stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var userID string
		var count int
		if err := rows.Scan(&userID, &count); err != nil {
			return nil, fmt.Errorf("scan reviewer stat: %w", err)
		}
		stats[userID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviewer stats: %w", err)
	}

	return stats, nil
}

func (r *PostgresPRRepository) GetPRStats(ctx context.Context) (map[string]int, error) {
	q := getQuerier(ctx, r.pool)

	query := `
		SELECT status, COUNT(*) as count
		FROM pull_requests
		GROUP BY status
	`

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query PR stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan PR stat: %w", err)
		}
		stats[status] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate PR stats: %w", err)
	}

	return stats, nil
}