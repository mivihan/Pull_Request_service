package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mivihan/Pull_Request_service/internal/domain"
)

type PostgresTeamRepository struct {
	pool *pgxpool.Pool
}

func NewTeamRepository(pool *pgxpool.Pool) TeamRepository {
	return &PostgresTeamRepository{pool: pool}
}

func (r *PostgresTeamRepository) Create(ctx context.Context, team *domain.Team) error {
	if err := team.Validate(); err != nil {
		return fmt.Errorf("invalid team: %w", err)
	}

	q := getQuerier(ctx, r.pool)

	query := `
		INSERT INTO teams (team_name, created_at)
		VALUES ($1, $2)
	`

	_, err := q.Exec(ctx, query, team.TeamName, team.CreatedAt)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrTeamExists
		}
		return fmt.Errorf("insert team: %w", err)
	}

	return nil
}

func (r *PostgresTeamRepository) GetByName(ctx context.Context, teamName string) (*domain.Team, error) {
	q := getQuerier(ctx, r.pool)

	query := `SELECT team_name, created_at FROM teams WHERE team_name = $1`

	var team domain.Team
	err := q.QueryRow(ctx, query, teamName).Scan(&team.TeamName, &team.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTeamNotFound
		}
		return nil, fmt.Errorf("query team: %w", err)
	}

	return &team, nil
}

func (r *PostgresTeamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	q := getQuerier(ctx, r.pool)

	query := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`

	var exists bool
	err := q.QueryRow(ctx, query, teamName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check team existence: %w", err)
	}

	return exists, nil
}

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
