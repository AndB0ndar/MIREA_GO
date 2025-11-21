package httpx

import (
	"github.com/go-chi/chi/v5"

	"app/internal/http/handlers"
	"app/internal/http/middleware"
)

func NewRouter(h *handlers.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.CORS)

	r.Route("/notes", func(r chi.Router) {
		r.Post("/", h.CreateNote)
		r.Get("/", h.GetAllNotes)
		r.Get("/{id}", h.GetNote)
		r.Patch("/{id}", h.UpdateNote)
		r.Delete("/{id}", h.DeleteNote)
	})

	return r
}
