package main

import (
	"app/internal/http"
	"app/internal/platform/config"
	"log"
	"net/http"
)

func main() {
	cfg := config.Load()

	mux := router.Build(cfg)

	log.Println("Server starting on", cfg.Port)
	log.Printf("Access token TTL: %v", cfg.AccessTokenTTL)
	log.Printf("Refresh token TTL: %v", cfg.RefreshTokenTTL)

	log.Fatal(http.ListenAndServe(cfg.Port, mux))
}
