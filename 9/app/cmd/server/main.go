package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

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

	if _, err := os.Stat(cfg.SocketPath); err == nil {
		log.Fatal("Error: socket file already exists at ", cfg.SocketPath)
	}
	listener, err := net.Listen("unix", cfg.SocketPath)
	if err != nil {
		log.Fatal("Failed to create Unix socket listener:", err)
	}
	defer func() {
		listener.Close()
		os.Remove(cfg.SocketPath)
	}()

	// Set appropriate permissions for the socket file
	if err := os.Chmod(cfg.SocketPath, 0666); err != nil {
		log.Fatal("Failed to set socket permissions:", err)
	}

	log.Println("Server starting on Unix socket:", cfg.SocketPath)
	log.Fatal(http.Serve(listener, r))
}
