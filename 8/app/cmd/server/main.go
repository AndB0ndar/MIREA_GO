package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"app/internal/db"
	"app/internal/notes"
)

func main() {
	socketPath := getEnv("SOCKET_PATH", "/tmp/app.sock")
	// Проверяем, существует ли socket файл
	if _, err := os.Stat(socketPath); err == nil {
		log.Fatalf("Error: socket file already exists at %s. Remove it manually or choose a different path.", socketPath)
	}

	// Подключаемся к MongoDB
	dsn := getEnv("MONGO_DSN", "mongodb://root:secret@localhost:27017/pz8?authSource=admin")
	ctx := context.Background()
	deps, err := db.ConnectMongo(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer deps.Disconnect(ctx)

	log.Println("Successfully connected to MongoDB!")

	// Создаем репозиторий и обработчики
	repo, err := notes.NewRepo(deps.Database)
	if err != nil {
		log.Fatalf("Failed to create notes repository: %v", err)
	}

	handler := notes.NewHandler(repo)

	// Настраиваем маршруты
	router := chi.NewRouter()

	// Middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			next.ServeHTTP(w, r)
		})
	})

	// Routes
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok","service":"notes-api"}`))
	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":"Notes API","version":"1.0.0"}`))
	})

	// Создаем listener на Unix socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to create Unix socket listener: %v", err)
	}

	// Устанавливаем права на socket файл
	if err := os.Chmod(socketPath, 0666); err != nil {
		log.Fatalf("Failed to set socket permissions: %v", err)
	}

	router.Mount("/notes", handler.Routes())

	// Запускаем сервер в отдельной goroutine
	server := &http.Server{
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on Unix socket: %s", socketPath)
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Ожидаем сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped")

	// Обеспечиваем очистку socket файла при завершении
	defer func() {
		listener.Close()
		os.Remove(socketPath)
	}()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
