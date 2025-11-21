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

// CreateNote godoc
// @Summary Создать заметку
// @Description Создает новую заметку
// @Tags notes
// @Accept json
// @Produce json
// @Param input body core.NoteCreate true "Данные новой заметки"
// @Success 201 {object} core.Note
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /notes [post]
func (h *Handler) CreateNote(w http.ResponseWriter, r *http.Request) {
	var req core.NoteCreate
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

// GetAllNotes godoc
// @Summary Получить все заметки
// @Description Возвращает список всех заметок
// @Tags notes
// @Produce json
// @Success 200 {array} core.Note
// @Failure 500 {object} map[string]string
// @Router /notes [get]
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

// GetNote godoc
// @Summary Получить заметку по ID
// @Description Возвращает заметку по указанному идентификатору
// @Tags notes
// @Produce json
// @Param id path int true "ID заметки"
// @Success 200 {object} core.Note
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /notes/{id} [get]
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

// UpdateNote godoc
// @Summary Обновить заметку
// @Description Обновляет существующую заметку (частичное обновление)
// @Tags notes
// @Accept json
// @Produce json
// @Param id path int true "ID заметки"
// @Param input body core.NoteUpdate true "Поля для обновления"
// @Success 200 {object} core.Note
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /notes/{id} [patch]
func (h *Handler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	var req core.NoteUpdate
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
		Title:   "",
		Content: "",
	}

	if req.Title != nil {
		updatedNote.Title = *req.Title
	}
	if req.Content != nil {
		updatedNote.Content = *req.Content
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

// DeleteNote godoc
// @Summary Удалить заметку
// @Description Удаляет заметку по указанному идентификатору
// @Tags notes
// @Param id path int true "ID заметки"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /notes/{id} [delete]
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
