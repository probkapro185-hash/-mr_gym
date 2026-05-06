package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"crm_gym/internal/config"
	"crm_gym/internal/dto"
	"crm_gym/internal/handler"
	"crm_gym/internal/repository"
	"crm_gym/internal/repository/postgres"
	"crm_gym/internal/service"
	jwtpkg "crm_gym/pkg/jwt"
	"crm_gym/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log := logger.New(os.Getenv("LOG_LEVEL"))

	cfg, err := config.Load()
	if err != nil {
		log.Error("load config", slog.Any("error", err))
		os.Exit(1)
	}

	// ─── Database ──────────────────────────────────────────────────────────
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DB.DSN())
	if err != nil {
		log.Error("connect to db", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Error("ping db", slog.Any("error", err))
		os.Exit(1)
	}
	log.Info("connected to database")

	// ─── JWT ───────────────────────────────────────────────────────────────
	jwtManager := jwtpkg.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)

	// ─── Repositories ──────────────────────────────────────────────────────
	userRepo := postgres.NewUserRepository(pool)
	sessionRepo := postgres.NewSessionRepository(pool)
	paymentRepo := postgres.NewPaymentRepository(pool)
	subRepo := postgres.NewSubscriptionRepository(pool)

	// ─── Services ──────────────────────────────────────────────────────────
	authSvc := service.NewAuthService(userRepo, jwtManager)
	userSvc := service.NewUserService(userRepo)
	sessionSvc := service.NewSessionService(sessionRepo, userRepo)
	paymentSvc := service.NewPaymentService(paymentRepo, userRepo)
	subSvc := service.NewSubscriptionService(subRepo, paymentRepo, userRepo)

	// ─── Handlers ──────────────────────────────────────────────────────────
	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	sessionH := handler.NewSessionHandler(sessionSvc)
	paymentH := handler.NewPaymentHandler(paymentSvc)
	subH := handler.NewSubscriptionHandler(subSvc)

	// ─── Ensure admin exists ───────────────────────────────────────────────
	ensureDefaultAdmin(context.Background(), userSvc, log)

	// ─── Router ────────────────────────────────────────────────────────────
	router := handler.NewRouter(jwtManager, authH, userH, sessionH, paymentH, subH)

	// ─── HTTP Server ───────────────────────────────────────────────────────
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ─── Graceful shutdown ─────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("server started", slog.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("listen and serve", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	<-quit
	log.Info("shutting down server...")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error("server shutdown", slog.Any("error", err))
	}
	log.Info("server stopped")

	_ = authSvc
}

// ensureDefaultAdmin создаёт первого администратора если пользователей нет
func ensureDefaultAdmin(ctx context.Context, userSvc *service.UserService, log *slog.Logger) {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminEmail == "" {
		adminEmail = "admin@gym.ru"
	}
	if adminPass == "" {
		adminPass = "Admin1234!"
	}

	_, total, err := userSvc.List(ctx, repository.UserFilter{Limit: 1})
	if err != nil || total > 0 {
		return
	}

	_, err = userSvc.CreateByAdmin(ctx, dto.CreateUserRequest{
		FullName: "Администратор",
		Phone:    "+70000000000",
		Email:    adminEmail,
		Password: adminPass,
		Role:     "admin",
	})
	if err != nil {
		log.Warn("could not create default admin", slog.Any("error", err))
		return
	}
	log.Info("default admin created",
		slog.String("email", adminEmail),
		slog.String("password", adminPass),
	)
}
