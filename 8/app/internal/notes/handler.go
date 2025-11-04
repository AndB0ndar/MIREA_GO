package notes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repo *Repo
}

func NewHandler(r *Repo) *Handler {
	return &Handler{repo: r}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Patch("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	return r
}

type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, Response{Error: err.Error()})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON"))
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, errors.New("title is required"))
		return
	}

	note, err := h.repo.Create(ctx, req.Title, req.Content)
	if err != nil {
		if errors.Is(err, ErrDuplicateKey) {
			writeError(w, http.StatusConflict, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, Response{Data: note})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := chi.URLParam(r, "id")
	note, err := h.repo.ByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound) || errors.Is(err, ErrInvalidID):
			writeError(w, http.StatusNotFound, errors.New("note not found"))
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, Response{Data: note})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	q := r.URL.Query().Get("q")

	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	skip, _ := strconv.ParseInt(r.URL.Query().Get("skip"), 10, 64)
	if skip < 0 {
		skip = 0
	}

	notes, err := h.repo.List(ctx, q, limit, skip)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, Response{Data: notes})
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := chi.URLParam(r, "id")

	var req UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON"))
		return
	}

	if req.Title == nil && req.Content == nil {
		writeError(w, http.StatusBadRequest, errors.New("no fields to update"))
		return
	}

	note, err := h.repo.Update(ctx, id, req.Title, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound) || errors.Is(err, ErrInvalidID):
			writeError(w, http.StatusNotFound, errors.New("note not found"))
		case errors.Is(err, ErrDuplicateKey):
			writeError(w, http.StatusConflict, err)
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, Response{Data: note})
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := chi.URLParam(r, "id")
	err := h.repo.Delete(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound) || errors.Is(err, ErrInvalidID):
			writeError(w, http.StatusNotFound, errors.New("note not found"))
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
