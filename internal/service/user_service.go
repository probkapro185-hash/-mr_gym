package service

import (
	"context"
	"fmt"

	"crm_gym/internal/domain"
	"crm_gym/internal/dto"
	"crm_gym/internal/repository"
	"crm_gym/internal/validator"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*dto.UserResponse, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toUserResponse(u), nil
}

func (s *UserService) GetMe(ctx context.Context, userID int64) (*dto.UserResponse, error) {
	return s.GetByID(ctx, userID)
}

// List — список пользователей (менеджер и админ)
func (s *UserService) List(ctx context.Context, f repository.UserFilter) ([]*dto.UserResponse, int, error) {
	users, total, err := s.userRepo.List(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]*dto.UserResponse, len(users))
	for i, u := range users {
		resp[i] = toUserResponse(u)
	}
	return resp, total, nil
}

// ApproveApplication — менеджер/админ одобряет заявку и устанавливает пароль
func (s *UserService) ApproveApplication(ctx context.Context, userID int64, req dto.SetPasswordRequest) error {
	if errs := validator.ValidatePassword(req.Password); errs != nil {
		return validator.ValidationErrors{*errs}
	}

	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.Status != domain.StatusPending {
		return fmt.Errorf("заявка уже обработана")
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	u.PasswordHash = hash
	u.Status = domain.StatusApproved
	return s.userRepo.Update(ctx, u)
}

// RejectApplication — менеджер/админ отклоняет заявку
func (s *UserService) RejectApplication(ctx context.Context, userID int64) error {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	u.Status = domain.StatusRejected
	return s.userRepo.Update(ctx, u)
}

// Update — обновление данных пользователя (менеджер/админ)
func (s *UserService) Update(ctx context.Context, userID int64, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.FullName != nil {
		u.FullName = *req.FullName
	}
	if req.Phone != nil {
		u.Phone = validator.NormalizePhone(*req.Phone)
	}
	if req.Email != nil {
		u.Email = *req.Email
	}
	if req.Notes != nil {
		u.Notes = *req.Notes
	}
	if req.TrainerID != nil {
		u.TrainerID = req.TrainerID
	}
	if req.Role != nil {
		u.Role = domain.Role(*req.Role)
	}
	if req.Status != nil {
		u.Status = domain.ApplicationStatus(*req.Status)
	}

	if err := s.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}
	return toUserResponse(u), nil
}

// CreateByAdmin — прямое создание пользователя администратором
func (s *UserService) CreateByAdmin(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	errs := validator.ValidateApplicationRequest(req.FullName, req.Phone, req.Email)
	if len(errs) > 0 {
		return nil, errs
	}
	if errs2 := validator.ValidatePassword(req.Password); errs2 != nil {
		return nil, validator.ValidationErrors{*errs2}
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	role := domain.Role(req.Role)
	if role == "" {
		role = domain.RoleClient
	}

	u := &domain.User{
		FullName:     req.FullName,
		Phone:        validator.NormalizePhone(req.Phone),
		Email:        req.Email,
		PasswordHash: hash,
		Role:         role,
		Status:       domain.StatusApproved,
		TrainerID:    req.TrainerID,
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}
	return toUserResponse(u), nil
}

// Delete — удаление пользователя (только админ)
func (s *UserService) Delete(ctx context.Context, userID int64) error {
	return s.userRepo.Delete(ctx, userID)
}

func toUserResponse(u *domain.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        u.ID,
		FullName:  u.FullName,
		Phone:     u.Phone,
		Email:     u.Email,
		Role:      string(u.Role),
		Status:    string(u.Status),
		Balance:   u.Balance,
		Visits:    u.Visits,
		Notes:     u.Notes,
		TrainerID: u.TrainerID,
		CreatedAt: u.CreatedAt,
	}
}
