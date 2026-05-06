package service

import (
	"context"
	"fmt"

	"crm_gym/internal/domain"
	"crm_gym/internal/dto"
	"crm_gym/internal/repository"
)

type SubscriptionService struct {
	subRepo     repository.SubscriptionRepository
	paymentRepo repository.PaymentRepository
	userRepo    repository.UserRepository
}

func NewSubscriptionService(
	subRepo repository.SubscriptionRepository,
	paymentRepo repository.PaymentRepository,
	userRepo repository.UserRepository,
) *SubscriptionService {
	return &SubscriptionService{subRepo: subRepo, paymentRepo: paymentRepo, userRepo: userRepo}
}

// Create — создание абонемента клиенту (менеджер/админ)
func (s *SubscriptionService) Create(ctx context.Context, req dto.CreateSubscriptionRequest) (*dto.SubscriptionResponse, error) {
	if req.Visits <= 0 {
		return nil, fmt.Errorf("количество посещений должно быть положительным")
	}
	if req.Price < 0 {
		return nil, fmt.Errorf("цена не может быть отрицательной")
	}

	u, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// Списываем с баланса, если цена > 0
	if req.Price > 0 {
		if u.Balance < req.Price {
			return nil, domain.ErrInsufficientBalance
		}
		if err := s.userRepo.UpdateBalance(ctx, req.UserID, -req.Price); err != nil {
			return nil, err
		}
		p := &domain.Payment{
			UserID:        req.UserID,
			Amount:        req.Price,
			ServiceName:   req.Name,
			OperationType: "subscription",
		}
		_ = s.paymentRepo.Create(ctx, p)
	}

	sub := &domain.Subscription{
		UserID:     req.UserID,
		Name:       req.Name,
		VisitsLeft: req.Visits,
		ValidUntil: req.ValidUntil,
		Price:      req.Price,
	}
	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}
	// Добавляем посещения
	_ = s.userRepo.UpdateVisits(ctx, req.UserID, req.Visits)

	return toSubResponse(sub), nil
}

func (s *SubscriptionService) ListByUser(ctx context.Context, userID int64) ([]*dto.SubscriptionResponse, error) {
	subs, err := s.subRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := make([]*dto.SubscriptionResponse, len(subs))
	for i, sub := range subs {
		resp[i] = toSubResponse(sub)
	}
	return resp, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, subID int64) error {
	return s.subRepo.Delete(ctx, subID)
}

func toSubResponse(s *domain.Subscription) *dto.SubscriptionResponse {
	return &dto.SubscriptionResponse{
		ID:         s.ID,
		UserID:     s.UserID,
		Name:       s.Name,
		VisitsLeft: s.VisitsLeft,
		ValidUntil: s.ValidUntil,
		Price:      s.Price,
		CreatedAt:  s.CreatedAt,
	}
}
