package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"app/internal/core"
	"app/internal/repo"
)

type Handler struct {
	Repo repo.NoteRepository
}

func (h *Handler) CreateNote(w http.ResponseWriter, r *http.Request) {
	var req core.NoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	note := core.Note{
		Title:   req.Title,
		Content: req.Content,
	}

	id, err := h.Repo.Create(note)
	if err != nil {
		http.Error(w, "Error creating note", http.StatusInternalServerError)
		return
	}

	createdNote, err := h.Repo.GetByID(id)
	if err != nil {
		http.Error(w, "Error retrieving created note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdNote)
}

func (h *Handler) GetAllNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := h.Repo.GetAll()
	if err != nil {
		http.Error(w, "Error retrieving notes", http.StatusInternalServerError)
		return
	}

	if notes == nil {
		notes = []*core.Note{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

func (h *Handler) GetNote(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	note, err := h.Repo.GetByID(id)
	if err != nil {
		if err == repo.ErrNoteNotFound {
			http.Error(w, "Note not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error retrieving note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

func (h *Handler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	var req core.NoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Проверяем существование заметки
	_, err = h.Repo.GetByID(id)
	if err != nil {
		if err == repo.ErrNoteNotFound {
			http.Error(w, "Note not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error retrieving note", http.StatusInternalServerError)
		return
	}

	updatedNote := core.Note{
		Title:   req.Title,
		Content: req.Content,
	}

	if err := h.Repo.Update(id, updatedNote); err != nil {
		http.Error(w, "Error updating note", http.StatusInternalServerError)
		return
	}

	// Возвращаем обновленную заметку
	note, err := h.Repo.GetByID(id)
	if err != nil {
		http.Error(w, "Error retrieving updated note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	if err := h.Repo.Delete(id); err != nil {
		if err == repo.ErrNoteNotFound {
			http.Error(w, "Note not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error deleting note", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
