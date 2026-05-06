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

type paymentRepo struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) repository.PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Create(ctx context.Context, p *domain.Payment) error {
	q := `INSERT INTO payments (user_id, amount, service_name, operation_type)
		  VALUES ($1,$2,$3,$4)
		  RETURNING id, created_at`
	err := r.db.QueryRow(ctx, q, p.UserID, p.Amount, p.ServiceName, p.OperationType).
		Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return fmt.Errorf("create payment: %w", err)
	}
	return nil
}

func (r *paymentRepo) GetByID(ctx context.Context, id int64) (*domain.Payment, error) {
	q := `SELECT p.id, p.user_id, u.full_name, p.amount, p.service_name, p.operation_type, p.created_at
		  FROM payments p JOIN users u ON u.id = p.user_id WHERE p.id = $1`
	p := &domain.Payment{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.UserID, &p.UserFullName, &p.Amount, &p.ServiceName, &p.OperationType, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get payment: %w", err)
	}
	return p, nil
}

func (r *paymentRepo) ListByUser(ctx context.Context, userID int64, from, to time.Time) ([]*domain.Payment, error) {
	q := `SELECT p.id, p.user_id, u.full_name, p.amount, p.service_name, p.operation_type, p.created_at
		  FROM payments p JOIN users u ON u.id = p.user_id
		  WHERE p.user_id = $1 AND p.created_at >= $2 AND p.created_at < $3
		  ORDER BY p.created_at DESC`
	rows, err := r.db.Query(ctx, q, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list payments by user: %w", err)
	}
	defer rows.Close()
	return scanPayments(rows)
}

func (r *paymentRepo) ListAll(ctx context.Context, from, to time.Time, limit, offset int) ([]*domain.Payment, int, error) {
	countQ := `SELECT COUNT(*) FROM payments WHERE created_at >= $1 AND created_at < $2`
	var total int
	if err := r.db.QueryRow(ctx, countQ, from, to).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count payments: %w", err)
	}

	if limit == 0 {
		limit = 50
	}

	q := `SELECT p.id, p.user_id, u.full_name, p.amount, p.service_name, p.operation_type, p.created_at
		  FROM payments p JOIN users u ON u.id = p.user_id
		  WHERE p.created_at >= $1 AND p.created_at < $2
		  ORDER BY p.created_at DESC LIMIT $3 OFFSET $4`
	rows, err := r.db.Query(ctx, q, from, to, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list all payments: %w", err)
	}
	defer rows.Close()
	payments, err := scanPayments(rows)
	return payments, total, err
}

func scanPayments(rows pgx.Rows) ([]*domain.Payment, error) {
	var payments []*domain.Payment
	for rows.Next() {
		p := &domain.Payment{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.UserFullName, &p.Amount, &p.ServiceName, &p.OperationType, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}
		payments = append(payments, p)
	}
	return payments, nil
}
