package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"app/internal/db"
	"app/internal/notes"
)

func main() {
	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Получаем настройки из переменных окружения
	dsn := getEnv("MONGO_DSN", "mongodb://root:secret@localhost:27017/pz8?authSource=admin")
	addr := getEnv("HTTP_ADDR", ":8080")

	// Подключаемся к MongoDB
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

	router.Mount("/notes", handler.Routes())

	// Запускаем сервер
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
