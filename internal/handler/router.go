package handler

import (
	"net/http"

	"crm_gym/internal/middleware"
	jwtpkg "crm_gym/pkg/jwt"
)

// NewRouter собирает все маршруты приложения.
// Используется стандартный net/http ServeMux (Go 1.22+ с поддержкой path params).
func NewRouter(
	jwtManager *jwtpkg.Manager,
	authH *AuthHandler,
	userH *UserHandler,
	sessionH *SessionHandler,
	paymentH *PaymentHandler,
	subH *SubscriptionHandler,
) http.Handler {
	mux := http.NewServeMux()

	// ─── Публичные маршруты ────────────────────────────────────────────────
	mux.HandleFunc("POST /api/v1/auth/apply", authH.Apply)
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authH.Refresh)

	// ─── Middleware ────────────────────────────────────────────────────────
	authMW := middleware.Auth(jwtManager)
	managerRoles := middleware.RequireRole("manager", "admin")
	adminOnly := middleware.RequireRole("admin")

	// ─── Вспомогательная функция обёртки ──────────────────────────────────
	chain := func(h http.HandlerFunc, mws ...func(http.Handler) http.Handler) http.Handler {
		var result http.Handler = h
		for i := len(mws) - 1; i >= 0; i-- {
			result = mws[i](result)
		}
		return result
	}

	// ─── Профиль (все авторизованные) ─────────────────────────────────────
	mux.Handle("GET /api/v1/users/me", chain(userH.GetMe, authMW))

	// ─── Расписание — клиент видит только своё ────────────────────────────
	mux.Handle("GET /api/v1/sessions/my", chain(sessionH.GetMySessions, authMW))

	// ─── Платежи клиента ──────────────────────────────────────────────────
	mux.Handle("GET /api/v1/payments/my", chain(paymentH.GetMyPayments, authMW))

	// ─── Абонементы клиента ───────────────────────────────────────────────
	mux.Handle("GET /api/v1/subscriptions/my", chain(subH.GetMy, authMW))

	// ─── Пользователи (менеджер + админ) ──────────────────────────────────
	mux.Handle("GET /api/v1/users", chain(userH.List, authMW, managerRoles))
	mux.Handle("GET /api/v1/users/{id}", chain(userH.GetByID, authMW, managerRoles))
	mux.Handle("PATCH /api/v1/users/{id}", chain(userH.Update, authMW, managerRoles))
	mux.Handle("POST /api/v1/users/{id}/approve", chain(userH.Approve, authMW, managerRoles))
	mux.Handle("POST /api/v1/users/{id}/reject", chain(userH.Reject, authMW, managerRoles))

	// ─── Пользователи (только админ) ──────────────────────────────────────
	mux.Handle("POST /api/v1/users", chain(userH.Create, authMW, adminOnly))
	mux.Handle("DELETE /api/v1/users/{id}", chain(userH.Delete, authMW, adminOnly))

	// ─── Расписание (менеджер + админ) ────────────────────────────────────
	mux.Handle("GET /api/v1/sessions", chain(sessionH.GetAll, authMW, managerRoles))
	mux.Handle("GET /api/v1/sessions/client/{id}", chain(sessionH.GetByClientID, authMW, managerRoles))
	mux.Handle("POST /api/v1/sessions", chain(sessionH.Create, authMW, managerRoles))
	mux.Handle("PATCH /api/v1/sessions/{id}", chain(sessionH.Update, authMW, managerRoles))
	mux.Handle("DELETE /api/v1/sessions/{id}", chain(sessionH.Delete, authMW, managerRoles))

	// ─── Платежи (менеджер/админ — пополнение; только админ — все/списание) ─
	mux.Handle("POST /api/v1/payments/deposit", chain(paymentH.Deposit, authMW, managerRoles))
	mux.Handle("GET /api/v1/payments/user/{id}", chain(paymentH.GetUserPayments, authMW, managerRoles))
	mux.Handle("GET /api/v1/payments", chain(paymentH.GetAll, authMW, adminOnly))
	mux.Handle("POST /api/v1/payments/charge", chain(paymentH.Charge, authMW, adminOnly))

	// ─── Абонементы (менеджер/админ — создание; только админ — удаление) ──
	mux.Handle("GET /api/v1/subscriptions/user/{id}", chain(subH.GetByUserID, authMW, managerRoles))
	mux.Handle("POST /api/v1/subscriptions", chain(subH.Create, authMW, managerRoles))
	mux.Handle("DELETE /api/v1/subscriptions/{id}", chain(subH.Delete, authMW, adminOnly))

	// ─── Health check ──────────────────────────────────────────────────────
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return middleware.CORS(mux)
}
