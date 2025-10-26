package httpapi

import (
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func BuildRouter(d *gorm.DB) *chi.Mux {
	r := chi.NewRouter()
	h := NewHandlers(d)

	prefix := GetAPIPrefix()
	if prefix != "" {
		r.Route(prefix, func(r chi.Router) { DefineRoutes(r, h) })
	} else {
		DefineRoutes(r, h)
	}

	return r
}

func DefineRoutes(r chi.Router, h *Handlers) {
	r.Get("/health", h.Health)
	r.Post("/users", h.CreateUser)
	r.Post("/notes", h.CreateNote)
	r.Get("/notes/{id}", h.GetNoteByID)
}

func GetAPIPrefix() string {
	prefix := os.Getenv("API_PREFIX")
	prefix = strings.TrimSpace(prefix)

	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	prefix = strings.TrimSuffix(prefix, "/")
	return prefix
}
