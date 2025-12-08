// Package main Notes API server.
//
// @title Notes API (PostgreSQL)
// @version 2.0
// @description Учебный REST API для управления заметками.
// @contact.name Andrey Bondar
// @contact.email andrey.bondar.2003@list.ru
// @BasePath /
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	httpx "app/internal/http"
	"app/internal/http/handlers"
	"app/internal/platform/config"
	"app/internal/platform/database"
	"app/internal/repo"
)

// @host localhost:8080
// @schemes http
func main() {
	// Инициализация логгера
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	// Загрузка конфигурации
	cfg := config.Load()
	sugar.Infow("Configuration loaded",
		"port", cfg.Port,
		"db_url", cfg.Database.URL[:min(len(cfg.Database.URL), 30)]+"...")

	// Инициализация базы данных
	if err := database.InitPostgres(cfg.Database); err != nil {
		sugar.Fatalw("Failed to initialize database", "error", err)
	}
	defer database.Close()

	// Репозиторий PostgreSQL
	noteRepo := repo.NewNoteRepoPG()
	handler := &handlers.Handler{Repo: noteRepo}
	router := httpx.NewRouter(handler)

	// Запуск сервера
	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sugar.Infow("Server starting",
			"address", srv.Addr,
			"max_conns", cfg.Database.MaxOpenConns,
			"idle_conns", cfg.Database.MaxIdleConns)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalw("Server failed to start", "error", err)
		}
	}()

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	sugar.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		sugar.Fatalw("Server forced to shutdown", "error", err)
	}

	sugar.Info("Server exited properly")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
