package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"app/internal/http"
	"app/internal/platform/config"
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

	// Set appropriate permissions for the socket file
	if err := os.Chmod(cfg.SocketPath, 0666); err != nil {
		log.Fatal("Failed to set socket permissions:", err)
	}

	mux := router.Build(cfg)

	log.Println("Server starting on Unix socket:", cfg.SocketPath)
	log.Printf("Access token TTL: %v", cfg.AccessTokenTTL)
	log.Printf("Refresh token TTL: %v", cfg.RefreshTokenTTL)

	log.Fatal(http.Serve(listener, mux))
}
