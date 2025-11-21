// Package main Notes API server.
//
// @title Notes API
// @version 1.0
// @description Учебный REST API для управления заметками (CRUD).
// @contact.name Andrey Bondar
// @contact.email andrey.bondar.2003@list.ru
// @BasePath /
package main

import (
	"log"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"

	httpx "app/internal/http"
	"app/internal/http/handlers"
	"app/internal/platform/config"
	"app/internal/repo"
)

// @host localhost:8080
// @schemes http
func main() {
	cfg := config.Load()

	noteRepo := repo.NewNoteRepoMem()
	handler := &handlers.Handler{Repo: noteRepo}
	router := httpx.NewRouter(handler)

	// router.Get("/docs/*", httpSwagger.WrapHandler)
	router.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/swagger.json"),
	))
	router.Get("/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.json")
	})

	log.Println("Server starting on", cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.Port, router))
}
