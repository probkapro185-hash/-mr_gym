package domain

import "time"

type Role string

const (
	RoleClient  Role = "client"
	RoleManager Role = "manager"
	RoleAdmin   Role = "admin"
)

// ApplicationStatus — статус заявки клиента
type ApplicationStatus string

const (
	StatusPending  ApplicationStatus = "pending"
	StatusApproved ApplicationStatus = "approved"
	StatusRejected ApplicationStatus = "rejected"
)

// User — основная сущность пользователя
type User struct {
	ID           int64             `json:"id"`
	FullName     string            `json:"full_name"`
	Phone        string            `json:"phone"`
	Email        string            `json:"email"`
	PasswordHash string            `json:"-"`
	Role         Role              `json:"role"`
	Status       ApplicationStatus `json:"status"`
	Balance      float64           `json:"balance"`
	Visits       int               `json:"visits"`
	Notes        string            `json:"notes"`      // "Мои данные" — заметки/доп. инфо
	TrainerID    *int64            `json:"trainer_id"` // закреплённый тренер
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// Trainer — тренер (тоже пользователь с ролью manager/admin, но отдельно для связей)
type Trainer struct {
	ID        int64    `json:"id"`
	FullName  string   `json:"full_name"`
	Phone     string   `json:"phone"`
	Email     string   `json:"email"`
	Specialty []string `json:"specialty"` // "Рельеф тела", "Похудение", "Набор массы"
	PhotoURL  string   `json:"photo_url"`
}

// Session — тренировка/занятие в расписании
type Session struct {
	ID          int64     `json:"id"`
	ClientID    int64     `json:"client_id"`
	TrainerID   int64     `json:"trainer_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status"` // "scheduled", "completed", "cancelled"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Payment — платёж
type Payment struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	UserFullName  string    `json:"user_full_name,omitempty"`
	Amount        float64   `json:"amount"`
	ServiceName   string    `json:"service_name"`   // название услуги
	OperationType string    `json:"operation_type"` // "deposit", "charge", "subscription"
	CreatedAt     time.Time `json:"created_at"`
}

// Subscription — абонемент
type Subscription struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Name       string    `json:"name"`
	VisitsLeft int       `json:"visits_left"`
	ValidUntil time.Time `json:"valid_until"`
	Price      float64   `json:"price"`
	CreatedAt  time.Time `json:"created_at"`
}
