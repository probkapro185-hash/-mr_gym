package service

import (
	"context"
	"fmt"

	"crm_gym/internal/domain"
	"crm_gym/internal/dto"
	"crm_gym/internal/repository"
	"crm_gym/internal/validator"
	jwtpkg "crm_gym/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo   repository.UserRepository
	jwtManager *jwtpkg.Manager
}

func NewAuthService(userRepo repository.UserRepository, jwtManager *jwtpkg.Manager) *AuthService {
	return &AuthService{userRepo: userRepo, jwtManager: jwtManager}
}

// SubmitApplication — клиент оставляет заявку (без пароля, статус pending)
func (s *AuthService) SubmitApplication(ctx context.Context, req dto.ApplicationRequest) error {
	errs := validator.ValidateApplicationRequest(req.FullName, req.Phone, req.Email)
	if len(errs) > 0 {
		return errs
	}

	phone := validator.NormalizePhone(req.Phone)

	// Проверяем дубли
	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return validator.ValidationErrors{{Field: "email", Message: "Пользователь с таким email уже существует"}}
	}
	if _, err := s.userRepo.GetByPhone(ctx, phone); err == nil {
		return validator.ValidationErrors{{Field: "phone", Message: "Пользователь с таким номером уже существует"}}
	}

	u := &domain.User{
		FullName: req.FullName,
		Phone:    phone,
		Email:    req.Email,
		Role:     domain.RoleClient,
		Status:   domain.StatusPending,
	}
	return s.userRepo.Create(ctx, u)
}

// Login — вход по email + пароль
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error) {
	u, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if u.Status == domain.StatusPending {
		return nil, fmt.Errorf("заявка ещё не рассмотрена")
	}
	if u.Status == domain.StatusRejected {
		return nil, fmt.Errorf("ваша заявка отклонена")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return s.generateTokens(u)
}

// Refresh — обновление пары токенов
func (s *AuthService) Refresh(ctx context.Context, req dto.RefreshRequest) (*dto.TokenResponse, error) {
	claims, err := s.jwtManager.Parse(req.RefreshToken)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	u, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return s.generateTokens(u)
}

func (s *AuthService) generateTokens(u *domain.User) (*dto.TokenResponse, error) {
	access, err := s.jwtManager.GenerateAccessToken(u.ID, string(u.Role))
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}
	refresh, err := s.jwtManager.GenerateRefreshToken(u.ID, string(u.Role))
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	return &dto.TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		Role:         string(u.Role),
		UserID:       u.ID,
	}, nil
}

// HashPassword — вспомогательная функция для хеширования пароля
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
