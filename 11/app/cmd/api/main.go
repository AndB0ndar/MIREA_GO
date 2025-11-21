package main

import (
	"log"
	"net/http"

	httpx "app/internal/http"
	"app/internal/http/handlers"
	"app/internal/platform/config"
	"app/internal/repo"
)

func main() {
	cfg := config.Load()

	noteRepo := repo.NewNoteRepoMem()
	handler := &handlers.Handler{Repo: noteRepo}
	router := httpx.NewRouter(handler)

	log.Println("Server starting on", cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.Port, router))
}
