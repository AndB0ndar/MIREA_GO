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
	"net"
	"net/http"
	"os"

	httpSwagger "github.com/swaggo/http-swagger"

	httpx "app/internal/http"
	"app/internal/http/handlers"
	"app/internal/platform/config"
	"app/internal/repo"
)

// @host arbond.ru/go/12
// @schemes http
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

	// router.Get("/docs/*", httpSwagger.WrapHandler)
	router.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("swagger.json"),
	))
	router.Get("/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.json")
	})

	log.Println("Server starting on Unix socket:", cfg.SocketPath)
	log.Fatal(http.Serve(listener, router))
}
