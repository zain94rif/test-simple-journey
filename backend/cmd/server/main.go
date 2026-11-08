package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	_ "helpdesk/backend/docs"
	"helpdesk/backend/internal/auth"
	"helpdesk/backend/internal/config"
	apphttp "helpdesk/backend/internal/http"
	"helpdesk/backend/internal/http/handler"
	"helpdesk/backend/internal/migrate"
	"helpdesk/backend/internal/repository/postgres"
	"helpdesk/backend/internal/seed"
	"helpdesk/backend/internal/service"
)

// @title           Helpdesk API
// @version         1.0
// @description     REST API Helpdesk technical test: autentikasi JWT, tiket dengan aturan transisi status, komentar publik/internal, dan manajemen pengguna oleh admin.
// @description     Semua endpoint kecuali health, register, dan login memerlukan header Authorization: Bearer <token>.
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Masukkan "Bearer" diikuti spasi dan JWT, contoh: Bearer eyJhbGciOi...
func main() {
	if err := run(); err != nil {
		slog.Error("aplikasi berhenti", "error", err)
		os.Exit(1)
	}
}

func run() error {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("konfigurasi: %w", err)
	}

	logger := newLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := connectWithRetry(ctx, cfg)
	if err != nil {
		return fmt.Errorf("koneksi database: %w", err)
	}
	defer pool.Close()

	if cfg.AutoMigrate {
		if err := migrate.Run(ctx, pool); err != nil {
			return fmt.Errorf("migrasi database: %w", err)
		}
		slog.Info("migrasi database selesai")
	}
	if cfg.AutoSeed {
		if err := seed.Run(ctx, pool, cfg); err != nil {
			return fmt.Errorf("seed database: %w", err)
		}
		slog.Info("seed database selesai")
	}

	userRepo := postgres.NewUserRepository(pool)
	ticketRepo := postgres.NewTicketRepository(pool)
	commentRepo := postgres.NewCommentRepository(pool)

	jwtManager := auth.NewManager(cfg.JWTSecret, cfg.JWTTTL)
	authService := service.NewAuthService(userRepo, jwtManager, cfg.BcryptCost)
	ticketService := service.NewTicketService(ticketRepo, commentRepo, userRepo)
	commentService := service.NewCommentService(ticketRepo, commentRepo)
	userService := service.NewUserService(userRepo)

	router := apphttp.NewRouter(apphttp.Dependencies{
		Config:   cfg,
		JWT:      jwtManager,
		Users:    userRepo,
		Auth:     handler.NewAuthHandler(authService),
		Tickets:  handler.NewTicketHandler(ticketService),
		Comments: handler.NewCommentHandler(commentService),
		UsersH:   handler.NewUserHandler(userService),
		Health:   handler.NewHealthHandler(pool),
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server berjalan", "port", cfg.Port, "env", cfg.AppEnv)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		slog.Info("sinyal berhenti diterima")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server HTTP: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	slog.Info("server berhenti dengan rapi")
	return nil
}

func connectWithRetry(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	var lastErr error
	for attempt := 1; attempt <= 30; attempt++ {
		pool, err := postgres.NewPool(ctx, cfg.DatabaseURL, cfg.DBMaxConns)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			err = pool.Ping(pingCtx)
			cancel()
			if err == nil {
				return pool, nil
			}
			pool.Close()
		}
		lastErr = err
		slog.Warn("database belum siap", "attempt", attempt, "error", err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return nil, lastErr
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}
