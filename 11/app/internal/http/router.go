package httpx

import (
    "github.com/go-chi/chi/v5"

    "app/internal/http/handlers"
)

func NewRouter(h *handlers.Handler) *chi.Mux {
    r := chi.NewRouter()

    r.Route("/notes", func(r chi.Router) {
        r.Post("/", h.CreateNote)
        r.Get("/", h.GetAllNotes)
        r.Get("/{id}", h.GetNote)
        r.Patch("/{id}", h.UpdateNote)
        r.Delete("/{id}", h.DeleteNote)
    })

    return r
}
