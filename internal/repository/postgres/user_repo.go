package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"crm_gym/internal/domain"
	"crm_gym/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, u *domain.User) error {
	q := `INSERT INTO users (full_name, phone, email, password_hash, role, status, balance, visits, notes, trainer_id)
		  VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		  RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, q,
		u.FullName, u.Phone, u.Email, u.PasswordHash,
		u.Role, u.Status, u.Balance, u.Visits, u.Notes, u.TrainerID,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	q := `SELECT id, full_name, phone, email, password_hash, role, status, balance, visits, notes, trainer_id, created_at, updated_at
		  FROM users WHERE id = $1`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.FullName, &u.Phone, &u.Email, &u.PasswordHash,
		&u.Role, &u.Status, &u.Balance, &u.Visits, &u.Notes, &u.TrainerID,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	q := `SELECT id, full_name, phone, email, password_hash, role, status, balance, visits, notes, trainer_id, created_at, updated_at
		  FROM users WHERE email = $1`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.FullName, &u.Phone, &u.Email, &u.PasswordHash,
		&u.Role, &u.Status, &u.Balance, &u.Visits, &u.Notes, &u.TrainerID,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

func (r *userRepo) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	q := `SELECT id, full_name, phone, email, password_hash, role, status, balance, visits, notes, trainer_id, created_at, updated_at
		  FROM users WHERE phone = $1`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, phone).Scan(
		&u.ID, &u.FullName, &u.Phone, &u.Email, &u.PasswordHash,
		&u.Role, &u.Status, &u.Balance, &u.Visits, &u.Notes, &u.TrainerID,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by phone: %w", err)
	}
	return u, nil
}

func (r *userRepo) List(ctx context.Context, f repository.UserFilter) ([]*domain.User, int, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	idx := 1

	if f.Role != "" {
		where = append(where, fmt.Sprintf("role = $%d", idx))
		args = append(args, f.Role)
		idx++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, f.Status)
		idx++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("(full_name ILIKE $%d OR email ILIKE $%d OR phone ILIKE $%d)", idx, idx+1, idx+2))
		like := "%" + f.Search + "%"
		args = append(args, like, like, like)
		idx += 3
	}

	whereClause := strings.Join(where, " AND ")

	countQ := "SELECT COUNT(*) FROM users WHERE " + whereClause
	var total int
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	limit := f.Limit
	if limit == 0 {
		limit = 50
	}
	args = append(args, limit, f.Offset)

	q := fmt.Sprintf(`SELECT id, full_name, phone, email, password_hash, role, status, balance, visits, notes, trainer_id, created_at, updated_at
		FROM users WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, whereClause, idx, idx+1)

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(
			&u.ID, &u.FullName, &u.Phone, &u.Email, &u.PasswordHash,
			&u.Role, &u.Status, &u.Balance, &u.Visits, &u.Notes, &u.TrainerID,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (r *userRepo) Update(ctx context.Context, u *domain.User) error {
	q := `UPDATE users SET full_name=$1, phone=$2, email=$3, password_hash=$4, role=$5,
		  status=$6, balance=$7, visits=$8, notes=$9, trainer_id=$10, updated_at=NOW()
		  WHERE id=$11
		  RETURNING updated_at`
	err := r.db.QueryRow(ctx, q,
		u.FullName, u.Phone, u.Email, u.PasswordHash,
		u.Role, u.Status, u.Balance, u.Visits, u.Notes, u.TrainerID,
		u.ID,
	).Scan(&u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (r *userRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.Exec(ctx, "DELETE FROM users WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *userRepo) UpdateBalance(ctx context.Context, userID int64, delta float64) error {
	q := `UPDATE users SET balance = balance + $1, updated_at = NOW() WHERE id = $2`
	res, err := r.db.Exec(ctx, q, delta, userID)
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *userRepo) UpdateVisits(ctx context.Context, userID int64, delta int) error {
	q := `UPDATE users SET visits = visits + $1, updated_at = NOW() WHERE id = $2`
	res, err := r.db.Exec(ctx, q, delta, userID)
	if err != nil {
		return fmt.Errorf("update visits: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
