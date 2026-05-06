package service

import (
	"context"
	"fmt"
	"time"

	"crm_gym/internal/domain"
	"crm_gym/internal/dto"
	"crm_gym/internal/repository"
)

type SessionService struct {
	sessionRepo repository.SessionRepository
	userRepo    repository.UserRepository
}

func NewSessionService(sessionRepo repository.SessionRepository, userRepo repository.UserRepository) *SessionService {
	return &SessionService{sessionRepo: sessionRepo, userRepo: userRepo}
}

func (s *SessionService) Create(ctx context.Context, req dto.CreateSessionRequest) (*dto.SessionResponse, error) {
	if req.StartTime.IsZero() || req.EndTime.IsZero() {
		return nil, fmt.Errorf("start_time и end_time обязательны")
	}
	if req.EndTime.Before(req.StartTime) {
		return nil, fmt.Errorf("end_time не может быть раньше start_time")
	}

	session := &domain.Session{
		ClientID:    req.ClientID,
		TrainerID:   req.TrainerID,
		Title:       req.Title,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Status:      "scheduled",
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}
	return s.enrichSession(ctx, session)
}

// GetForClient — получение занятий конкретного клиента (клиент видит только свои)
func (s *SessionService) GetForClient(ctx context.Context, clientID int64, from, to time.Time) ([]*dto.SessionResponse, error) {
	sessions, err := s.sessionRepo.ListByClient(ctx, clientID, from, to)
	if err != nil {
		return nil, err
	}
	return s.enrichSessions(ctx, sessions)
}

// GetAll — все занятия (менеджер/админ)
func (s *SessionService) GetAll(ctx context.Context, from, to time.Time) ([]*dto.SessionResponse, error) {
	sessions, err := s.sessionRepo.ListAll(ctx, from, to)
	if err != nil {
		return nil, err
	}
	return s.enrichSessions(ctx, sessions)
}

// Update — изменение занятия (время, дата, тренер) — менеджер/админ
func (s *SessionService) Update(ctx context.Context, sessionID int64, req dto.UpdateSessionRequest) (*dto.SessionResponse, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if req.TrainerID != nil {
		session.TrainerID = *req.TrainerID
	}
	if req.Title != nil {
		session.Title = *req.Title
	}
	if req.Description != nil {
		session.Description = *req.Description
	}
	if req.StartTime != nil {
		session.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		session.EndTime = *req.EndTime
	}
	if req.Status != nil {
		session.Status = *req.Status
	}

	if session.EndTime.Before(session.StartTime) {
		return nil, fmt.Errorf("end_time не может быть раньше start_time")
	}

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, err
	}
	return s.enrichSession(ctx, session)
}

// Delete — удаление занятия
func (s *SessionService) Delete(ctx context.Context, sessionID int64) error {
	return s.sessionRepo.Delete(ctx, sessionID)
}

func (s *SessionService) enrichSession(ctx context.Context, session *domain.Session) (*dto.SessionResponse, error) {
	resp := toSessionResponse(session)

	if client, err := s.userRepo.GetByID(ctx, session.ClientID); err == nil {
		resp.ClientName = client.FullName
	}
	if trainer, err := s.userRepo.GetByID(ctx, session.TrainerID); err == nil {
		resp.TrainerName = trainer.FullName
	}
	return resp, nil
}

func (s *SessionService) enrichSessions(ctx context.Context, sessions []*domain.Session) ([]*dto.SessionResponse, error) {
	resp := make([]*dto.SessionResponse, 0, len(sessions))
	for _, session := range sessions {
		r, err := s.enrichSession(ctx, session)
		if err != nil {
			return nil, err
		}
		resp = append(resp, r)
	}
	return resp, nil
}

func toSessionResponse(s *domain.Session) *dto.SessionResponse {
	return &dto.SessionResponse{
		ID:          s.ID,
		ClientID:    s.ClientID,
		TrainerID:   s.TrainerID,
		Title:       s.Title,
		Description: s.Description,
		StartTime:   s.StartTime,
		EndTime:     s.EndTime,
		Status:      s.Status,
	}
}
