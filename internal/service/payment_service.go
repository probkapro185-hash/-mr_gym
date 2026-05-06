package service

import (
	"context"
	"fmt"
	"time"

	"crm_gym/internal/domain"
	"crm_gym/internal/dto"
	"crm_gym/internal/repository"
)

type PaymentService struct {
	paymentRepo repository.PaymentRepository
	userRepo    repository.UserRepository
}

func NewPaymentService(paymentRepo repository.PaymentRepository, userRepo repository.UserRepository) *PaymentService {
	return &PaymentService{paymentRepo: paymentRepo, userRepo: userRepo}
}

// Deposit — пополнение баланса клиента (менеджер/админ)
func (s *PaymentService) Deposit(ctx context.Context, req dto.DepositRequest) (*dto.PaymentResponse, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("сумма должна быть положительной")
	}

	if _, err := s.userRepo.GetByID(ctx, req.UserID); err != nil {
		return nil, fmt.Errorf("пользователь не найден: %w", err)
	}

	// Начисляем на баланс
	if err := s.userRepo.UpdateBalance(ctx, req.UserID, req.Amount); err != nil {
		return nil, err
	}

	serviceName := req.ServiceName
	if serviceName == "" {
		serviceName = "Пополнение баланса"
	}

	p := &domain.Payment{
		UserID:        req.UserID,
		Amount:        req.Amount,
		ServiceName:   serviceName,
		OperationType: "deposit",
	}
	if err := s.paymentRepo.Create(ctx, p); err != nil {
		return nil, err
	}

	full, _ := s.paymentRepo.GetByID(ctx, p.ID)
	return toPaymentResponse(full), nil
}

// Charge — списание с баланса (менеджер/админ)
func (s *PaymentService) Charge(ctx context.Context, req dto.ChargeRequest) (*dto.PaymentResponse, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("сумма должна быть положительной")
	}

	u, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if u.Balance < req.Amount {
		return nil, domain.ErrInsufficientBalance
	}

	if err := s.userRepo.UpdateBalance(ctx, req.UserID, -req.Amount); err != nil {
		return nil, err
	}

	p := &domain.Payment{
		UserID:        req.UserID,
		Amount:        req.Amount,
		ServiceName:   req.ServiceName,
		OperationType: "charge",
	}
	if err := s.paymentRepo.Create(ctx, p); err != nil {
		return nil, err
	}

	full, _ := s.paymentRepo.GetByID(ctx, p.ID)
	return toPaymentResponse(full), nil
}

// ListByUser — история платежей клиента
func (s *PaymentService) ListByUser(ctx context.Context, userID int64, from, to time.Time) ([]*dto.PaymentResponse, error) {
	payments, err := s.paymentRepo.ListByUser(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
	return toPaymentResponses(payments), nil
}

// ListAll — все финансы (только для админа)
func (s *PaymentService) ListAll(ctx context.Context, from, to time.Time, limit, offset int) ([]*dto.PaymentResponse, int, error) {
	payments, total, err := s.paymentRepo.ListAll(ctx, from, to, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return toPaymentResponses(payments), total, nil
}

func toPaymentResponse(p *domain.Payment) *dto.PaymentResponse {
	if p == nil {
		return nil
	}
	return &dto.PaymentResponse{
		ID:            p.ID,
		UserID:        p.UserID,
		UserFullName:  p.UserFullName,
		Amount:        p.Amount,
		ServiceName:   p.ServiceName,
		OperationType: p.OperationType,
		CreatedAt:     p.CreatedAt,
	}
}

func toPaymentResponses(payments []*domain.Payment) []*dto.PaymentResponse {
	resp := make([]*dto.PaymentResponse, len(payments))
	for i, p := range payments {
		resp[i] = toPaymentResponse(p)
	}
	return resp
}
