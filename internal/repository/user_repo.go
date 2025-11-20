package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mivihan/Pull_Request_service/internal/domain"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Upsert(ctx context.Context, user *domain.User) error {
	if err := user.Validate(); err != nil {
		return fmt.Errorf("invalid user: %w", err)
	}

	q := getQuerier(ctx, r.pool)

	query := `
		INSERT INTO users (user_id, username, team_name, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) 
		DO UPDATE SET
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active
	`

	_, err := q.Exec(ctx, query,
		user.UserID,
		user.Username,
		user.TeamName,
		user.IsActive,
		user.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	q := getQuerier(ctx, r.pool)

	query := `
		SELECT user_id, username, team_name, is_active, created_at
		FROM users
		WHERE user_id = $1
	`

	var user domain.User
	err := q.QueryRow(ctx, query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	return &user, nil
}

func (r *PostgresUserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	q := getQuerier(ctx, r.pool)

	query := `
		UPDATE users
		SET is_active = $2
		WHERE user_id = $1
		RETURNING user_id, username, team_name, is_active, created_at
	`

	var user domain.User
	err := q.QueryRow(ctx, query, userID, isActive).Scan(
		&user.UserID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("update user is_active: %w", err)
	}

	return &user, nil
}

func (r *PostgresUserRepository) ListByTeam(ctx context.Context, teamName string) ([]*domain.User, error) {
	q := getQuerier(ctx, r.pool)

	query := `
		SELECT user_id, username, team_name, is_active, created_at
		FROM users
		WHERE team_name = $1
		ORDER BY user_id
	`

	rows, err := q.Query(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("query users by team: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.TeamName,
			&user.IsActive,
			&user.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}

func (r *PostgresUserRepository) ListActiveByTeamExcluding(
	ctx context.Context,
	teamName string,
	excludeUserIDs []string,
) ([]*domain.User, error) {
	q := getQuerier(ctx, r.pool)
	query := `
		SELECT user_id, username, team_name, is_active, created_at
		FROM users
		WHERE team_name = $1 
		  AND is_active = true
	`

	args := []interface{}{teamName}

	if len(excludeUserIDs) > 0 {
		placeholders := make([]string, len(excludeUserIDs))
		for i, id := range excludeUserIDs {
			args = append(args, id)
			placeholders[i] = fmt.Sprintf("$%d", i+2)
		}
		query += fmt.Sprintf(" AND user_id NOT IN (%s)", strings.Join(placeholders, ", "))
	}

	query += " ORDER BY user_id"

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query active users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.TeamName,
			&user.IsActive,
			&user.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}
