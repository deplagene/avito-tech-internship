package main

import (
	"context"
	"deplagene/avito-tech-internship/cmd/api"
	"deplagene/avito-tech-internship/configs"
	"deplagene/avito-tech-internship/db"
	"deplagene/avito-tech-internship/internal/pullrequest"
	"deplagene/avito-tech-internship/internal/team"
	"deplagene/avito-tech-internship/internal/user"
	"deplagene/avito-tech-internship/utils"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Загружаем конфиг
	cfg := configs.InitConfig()

	// Инициализируем логгер
	logger := utils.NewLogger(cfg.Env)
	slog.SetDefault(logger) // Set default logger for convenience

	logger.Info("Starting application", "env", cfg.Env, "port", cfg.HTTPPort)

	// Подключаемся к БД
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("Connected to database")

	// Инициализируем репозитории
	teamRepo := team.NewTeamRepository(pool)
	userRepo := user.NewUserRepository(pool)
	prRepo := pullrequest.NewPullRequestRepository(pool)

	// Инициализируем сервисы
	teamService := team.NewService(teamRepo, userRepo, pool, logger)
	userService := user.NewService(userRepo, pool, logger)
	prService := pullrequest.NewService(prRepo, userRepo, teamRepo, pool, logger)

	// Создаем хендлер
	apiHandler := pullrequest.NewHandler(teamService, userService, prService, logger)

	// Настройка роутера Chi
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Регистрируем роуты, который сгнерерил oapi-codegen
	api.HandlerWithOptions(apiHandler, api.ChiServerOptions{BaseRouter: router})

	// Дополнительные роуты (health-check)
	router.Get("/health", apiHandler.GetHealth)

	// Запускаем сервер
	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutting down server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Server shutdown failed", "error", err)
		}
	}()

	logger.Info("Server listening", "port", cfg.HTTPPort)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
	logger.Info("Server stopped")
}
