package main

import (
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"

	"app/internal/http/handlers"
	"app/internal/platform/config"
	"app/internal/repo"
)

func main() {
	cfg := config.Load()

	db, err := repo.Open(cfg.DB_DSN)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	userRepo := repo.NewUserRepo(db)

	if err := userRepo.AutoMigrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	authHandler := &handlers.AuthHandler{
		Users:      userRepo,
		BcryptCost: cfg.BcryptCost,
	}

	r := chi.NewRouter()
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	log.Println("Server starting on", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, r))
}
