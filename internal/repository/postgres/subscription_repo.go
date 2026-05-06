package postgres

import (
	"context"
	"errors"
	"fmt"

	"crm_gym/internal/domain"
	"crm_gym/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type subscriptionRepo struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) repository.SubscriptionRepository {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) Create(ctx context.Context, s *domain.Subscription) error {
	q := `INSERT INTO subscriptions (user_id, name, visits_left, valid_until, price)
		  VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`
	return r.db.QueryRow(ctx, q, s.UserID, s.Name, s.VisitsLeft, s.ValidUntil, s.Price).
		Scan(&s.ID, &s.CreatedAt)
}

func (r *subscriptionRepo) GetByID(ctx context.Context, id int64) (*domain.Subscription, error) {
	q := `SELECT id, user_id, name, visits_left, valid_until, price, created_at FROM subscriptions WHERE id=$1`
	s := &domain.Subscription{}
	err := r.db.QueryRow(ctx, q, id).Scan(&s.ID, &s.UserID, &s.Name, &s.VisitsLeft, &s.ValidUntil, &s.Price, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	return s, nil
}

func (r *subscriptionRepo) ListByUser(ctx context.Context, userID int64) ([]*domain.Subscription, error) {
	q := `SELECT id, user_id, name, visits_left, valid_until, price, created_at FROM subscriptions WHERE user_id=$1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()
	var subs []*domain.Subscription
	for rows.Next() {
		s := &domain.Subscription{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.VisitsLeft, &s.ValidUntil, &s.Price, &s.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *subscriptionRepo) Update(ctx context.Context, s *domain.Subscription) error {
	q := `UPDATE subscriptions SET name=$1, visits_left=$2, valid_until=$3, price=$4 WHERE id=$5`
	res, err := r.db.Exec(ctx, q, s.Name, s.VisitsLeft, s.ValidUntil, s.Price, s.ID)
	if err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *subscriptionRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.Exec(ctx, "DELETE FROM subscriptions WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
