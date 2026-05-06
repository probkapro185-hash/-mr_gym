package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"crm_gym/internal/domain"
	"crm_gym/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sessionRepo struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) repository.SessionRepository {
	return &sessionRepo{db: db}
}

const sessionSelect = `SELECT id, client_id, trainer_id, title, description, start_time, end_time, status, created_at, updated_at FROM sessions`

func scanSession(row pgx.Row) (*domain.Session, error) {
	s := &domain.Session{}
	err := row.Scan(&s.ID, &s.ClientID, &s.TrainerID, &s.Title, &s.Description,
		&s.StartTime, &s.EndTime, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return s, nil
}

func (r *sessionRepo) Create(ctx context.Context, s *domain.Session) error {
	q := `INSERT INTO sessions (client_id, trainer_id, title, description, start_time, end_time, status)
		  VALUES ($1,$2,$3,$4,$5,$6,$7)
		  RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, q,
		s.ClientID, s.TrainerID, s.Title, s.Description, s.StartTime, s.EndTime, s.Status,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *sessionRepo) GetByID(ctx context.Context, id int64) (*domain.Session, error) {
	q := sessionSelect + ` WHERE id = $1`
	s, err := scanSession(r.db.QueryRow(ctx, q, id))
	if err != nil {
		return nil, fmt.Errorf("get session by id: %w", err)
	}
	return s, nil
}

func (r *sessionRepo) listSessions(ctx context.Context, where string, args []interface{}) ([]*domain.Session, error) {
	q := sessionSelect + ` WHERE ` + where + ` ORDER BY start_time ASC`
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		s := &domain.Session{}
		if err := rows.Scan(&s.ID, &s.ClientID, &s.TrainerID, &s.Title, &s.Description,
			&s.StartTime, &s.EndTime, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *sessionRepo) ListByClient(ctx context.Context, clientID int64, from, to time.Time) ([]*domain.Session, error) {
	return r.listSessions(ctx, "client_id = $1 AND start_time >= $2 AND start_time < $3",
		[]interface{}{clientID, from, to})
}

func (r *sessionRepo) ListByTrainer(ctx context.Context, trainerID int64, from, to time.Time) ([]*domain.Session, error) {
	return r.listSessions(ctx, "trainer_id = $1 AND start_time >= $2 AND start_time < $3",
		[]interface{}{trainerID, from, to})
}

func (r *sessionRepo) ListAll(ctx context.Context, from, to time.Time) ([]*domain.Session, error) {
	return r.listSessions(ctx, "start_time >= $1 AND start_time < $2",
		[]interface{}{from, to})
}

func (r *sessionRepo) Update(ctx context.Context, s *domain.Session) error {
	q := `UPDATE sessions SET trainer_id=$1, title=$2, description=$3, start_time=$4, end_time=$5, status=$6, updated_at=NOW()
		  WHERE id=$7 RETURNING updated_at`
	err := r.db.QueryRow(ctx, q,
		s.TrainerID, s.Title, s.Description, s.StartTime, s.EndTime, s.Status, s.ID,
	).Scan(&s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("update session: %w", err)
	}
	return nil
}

func (r *sessionRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.Exec(ctx, "DELETE FROM sessions WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
