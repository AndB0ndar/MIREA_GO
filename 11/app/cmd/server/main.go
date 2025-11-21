package main

import (
	"log"
	"net"
	"net/http"
	"os"

	httpx "app/internal/http"
	"app/internal/http/handlers"
	"app/internal/platform/config"
	"app/internal/repo"
)

func main() {
	cfg := config.Load()

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
	if err := os.Chmod(cfg.SocketPath, 0666); err != nil {
		log.Fatal("Failed to set socket permissions:", err)
	}

	noteRepo := repo.NewNoteRepoMem()
	handler := &handlers.Handler{Repo: noteRepo}
	router := httpx.NewRouter(handler)

	log.Println("Server starting on Unix socket:", cfg.SocketPath)
	log.Fatal(http.Serve(listener, router))
}
