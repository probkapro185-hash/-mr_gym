package repository

import (
	"context"
	"crm_gym/internal/domain"
	"time"
)

// UserRepository — работа с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)
	List(ctx context.Context, filter UserFilter) ([]*domain.User, int, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id int64) error
	UpdateBalance(ctx context.Context, userID int64, delta float64) error
	UpdateVisits(ctx context.Context, userID int64, delta int) error
}

type UserFilter struct {
	Role   string
	Status string
	Search string
	Limit  int
	Offset int
}

// SessionRepository — работа с расписанием
type SessionRepository interface {
	Create(ctx context.Context, s *domain.Session) error
	GetByID(ctx context.Context, id int64) (*domain.Session, error)
	ListByClient(ctx context.Context, clientID int64, from, to time.Time) ([]*domain.Session, error)
	ListByTrainer(ctx context.Context, trainerID int64, from, to time.Time) ([]*domain.Session, error)
	ListAll(ctx context.Context, from, to time.Time) ([]*domain.Session, error)
	Update(ctx context.Context, s *domain.Session) error
	Delete(ctx context.Context, id int64) error
}

// PaymentRepository — работа с финансами
type PaymentRepository interface {
	Create(ctx context.Context, p *domain.Payment) error
	GetByID(ctx context.Context, id int64) (*domain.Payment, error)
	ListByUser(ctx context.Context, userID int64, from, to time.Time) ([]*domain.Payment, error)
	ListAll(ctx context.Context, from, to time.Time, limit, offset int) ([]*domain.Payment, int, error)
}

// SubscriptionRepository — абонементы
type SubscriptionRepository interface {
	Create(ctx context.Context, s *domain.Subscription) error
	GetByID(ctx context.Context, id int64) (*domain.Subscription, error)
	ListByUser(ctx context.Context, userID int64) ([]*domain.Subscription, error)
	Update(ctx context.Context, s *domain.Subscription) error
	Delete(ctx context.Context, id int64) error
}
