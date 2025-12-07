package httpx

import (
    "context"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    httpSwagger "github.com/swaggo/http-swagger"

    "app/internal/http/handlers"
    "app/internal/http/middleware"
    "app/internal/platform/database"
)

func NewRouter(h *handlers.Handler) *chi.Mux {
    r := chi.NewRouter()

    r.Use(middleware.CORS)
    r.Use(middleware.Logger) // Добавьте middleware для логирования

    r.Route("/notes", func(r chi.Router) {
        r.Post("/", h.CreateNote)
        r.Get("/", h.GetAllNotes)                // Старый метод без пагинации
        r.Get("/offset", h.GetNotesOffset)       // OFFSET пагинация (неоптимизированная)
        r.Get("/paginated", h.GetNotesPaginated) // Keyset-пагинация
        r.Get("/batch", h.GetNotesBatch)         // Батчинг
        r.Get("/search", h.SearchNotes)          // Поиск
        r.Get("/explain", h.ExplainQuery)        // EXPLAIN ANALYZE
        r.Get("/{id}", h.GetNote)
        r.Patch("/{id}", h.UpdateNote)
        r.Delete("/{id}", h.DeleteNote)
    })

    // Swagger
    r.Get("/docs/*", httpSwagger.Handler(httpSwagger.URL("swagger.json")))
    r.Get("/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "docs/swagger.json")
    })

    // Статистика БД и здоровья
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()
        if err := database.Pool.Ping(ctx); err != nil {
            http.Error(w, `{"status": "down", "database": "unavailable"}`, http.StatusServiceUnavailable)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"status": "up", "database": "connected"}`))
    })
    r.Get("/api/stats/db", func(w http.ResponseWriter, r *http.Request) {
        stats := database.GetStats()
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"status": "ok", "pool_stats": "` + stats + `"}`))
    })

    return r
}
