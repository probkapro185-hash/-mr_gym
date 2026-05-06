package dto

import "time"

// --- Auth / Registration ---

// ApplicationRequest — заявка от клиента (вкладка "Войти/Записаться")
type ApplicationRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

// LoginRequest — вход по email + пароль
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenResponse — ответ при успешной аутентификации
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Role         string `json:"role"`
	UserID       int64  `json:"user_id"`
}

// RefreshRequest — обновление токена
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// --- User ---

type UserResponse struct {
	ID        int64     `json:"id"`
	FullName  string    `json:"full_name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	Balance   float64   `json:"balance"`
	Visits    int       `json:"visits"`
	Notes     string    `json:"notes"`
	TrainerID *int64    `json:"trainer_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// UpdateUserRequest — обновление данных пользователя (менеджер/админ)
type UpdateUserRequest struct {
	FullName  *string `json:"full_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Email     *string `json:"email,omitempty"`
	Notes     *string `json:"notes,omitempty"`
	TrainerID *int64  `json:"trainer_id,omitempty"`
	Role      *string `json:"role,omitempty"`
	Status    *string `json:"status,omitempty"`
}

// SetPasswordRequest — установка пароля (менеджер/админ при одобрении заявки)
type SetPasswordRequest struct {
	Password string `json:"password"`
}

// CreateUserRequest — прямое создание пользователя (только админ)
type CreateUserRequest struct {
	FullName  string `json:"full_name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Role      string `json:"role"`
	TrainerID *int64 `json:"trainer_id,omitempty"`
}

// --- Sessions (Расписание) ---

type SessionResponse struct {
	ID          int64     `json:"id"`
	ClientID    int64     `json:"client_id"`
	ClientName  string    `json:"client_name,omitempty"`
	TrainerID   int64     `json:"trainer_id"`
	TrainerName string    `json:"trainer_name,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status"`
}

type CreateSessionRequest struct {
	ClientID    int64     `json:"client_id"`
	TrainerID   int64     `json:"trainer_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
}

type UpdateSessionRequest struct {
	TrainerID   *int64     `json:"trainer_id,omitempty"`
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Status      *string    `json:"status,omitempty"`
}

// --- Payments (Финансы) ---

type PaymentResponse struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	UserFullName  string    `json:"user_full_name"`
	Amount        float64   `json:"amount"`
	ServiceName   string    `json:"service_name"`
	OperationType string    `json:"operation_type"`
	CreatedAt     time.Time `json:"created_at"`
}

type DepositRequest struct {
	UserID      int64   `json:"user_id"`
	Amount      float64 `json:"amount"`
	ServiceName string  `json:"service_name"`
}

type ChargeRequest struct {
	UserID      int64   `json:"user_id"`
	Amount      float64 `json:"amount"`
	ServiceName string  `json:"service_name"`
}

// --- Subscriptions ---

type SubscriptionResponse struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Name       string    `json:"name"`
	VisitsLeft int       `json:"visits_left"`
	ValidUntil time.Time `json:"valid_until"`
	Price      float64   `json:"price"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateSubscriptionRequest struct {
	UserID     int64     `json:"user_id"`
	Name       string    `json:"name"`
	Visits     int       `json:"visits"`
	ValidUntil time.Time `json:"valid_until"`
	Price      float64   `json:"price"`
}

// --- Trainers ---

type TrainerResponse struct {
	ID        int64    `json:"id"`
	FullName  string   `json:"full_name"`
	Phone     string   `json:"phone"`
	Email     string   `json:"email"`
	Specialty []string `json:"specialty"`
	PhotoURL  string   `json:"photo_url"`
}

// --- Generic ---

type ErrorResponse struct {
	Error   string      `json:"error"`
	Details interface{} `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ListResponse struct {
	Items interface{} `json:"items"`
	Total int         `json:"total"`
}
