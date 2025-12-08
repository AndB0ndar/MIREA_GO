package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

// GetNotesPaginated godoc
// @Summary Получить заметки с keyset-пагинацией
// @Description Возвращает заметки с использованием keyset-пагинации
// @Tags notes
// @Produce json
// @Param last_id query int false "ID последней заметки на предыдущей странице"
// @Param limit query int false "Количество записей" default(10)
// @Success 200 {array} core.Note
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /notes/paginated [get]
func (h *Handler) GetNotesPaginated(w http.ResponseWriter, r *http.Request) {
	lastID, _ := strconv.ParseInt(r.URL.Query().Get("last_id"), 10, 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	notes, err := h.Repo.GetPaginated(lastID, limit)
	if err != nil {
		http.Error(w, "Error retrieving notes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

// GetNotesBatch godoc
// @Summary Получить несколько заметок за один запрос
// @Description Возвращает заметки по списку ID (батчинг)
// @Tags notes
// @Produce json
// @Param ids query string true "Список ID через запятую"
// @Success 200 {array} core.Note
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /notes/batch [get]
func (h *Handler) GetNotesBatch(w http.ResponseWriter, r *http.Request) {
	idsParam := r.URL.Query().Get("ids")
	if idsParam == "" {
		http.Error(w, "IDs parameter is required", http.StatusBadRequest)
		return
	}

	idStrs := strings.Split(idsParam, ",")
	var ids []int64
	for _, idStr := range idStrs {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}
		ids = append(ids, id)
	}

	notes, err := h.Repo.GetBatch(ids)
	if err != nil {
		http.Error(w, "Error retrieving notes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

// SearchNotes godoc
// @Summary Поиск заметок по заголовку
// @Description Поиск с использованием полнотекстового индекса
// @Tags notes
// @Produce json
// @Param q query string true "Поисковый запрос"
// @Param limit query int false "Количество результатов" default(10)
// @Success 200 {array} core.Note
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /notes/search [get]
func (h *Handler) SearchNotes(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Search query is required", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	notes, err := h.Repo.SearchByTitle(query, limit)
	if err != nil {
		http.Error(w, "Error searching notes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

// GetNotesOffset godoc
// @Summary Получить заметки с OFFSET пагинацией
// @Description Возвращает заметки с использованием OFFSET/LIMIT (неоптимизированно)
// @Tags notes
// @Produce json
// @Param limit query int false "Количество записей" default(10)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {object} core.SuccessResponse
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notes/offset [get]
func (h *Handler) GetNotesOffset(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Используем новый метод с пагинацией
	if repoPG, ok := h.Repo.(interface {
		GetAllPaginated(limit, offset int) ([]*core.Note, int, error)
	}); ok {
		notes, total, err := repoPG.GetAllPaginated(limit, offset)
		if err != nil {
			http.Error(w, "Error retrieving notes", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"notes":  notes,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		// Fallback для in-memory репозитория
		http.Error(w, "Offset pagination not supported", http.StatusNotImplemented)
	}
}

// ExplainQuery godoc
// @Summary Выполнить EXPLAIN ANALYZE для запроса
// @Description Показать план выполнения SQL запроса (только для администраторов)
// @Tags notes
// @Produce json
// @Param query query string true "SQL запрос для анализа"
// @Success 200 {object} core.ExplainResponse
// @Failure 400 {object} core.ErrorResponse
// @Failure 500 {object} core.ErrorResponse
// @Router /notes/explain [get]
func (h *Handler) ExplainQuery(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	if repoPG, ok := h.Repo.(interface {
		ExplainQuery(query string, params ...interface{}) (string, error)
	}); ok {
		explanation, err := repoPG.ExplainQuery(query)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error explaining query: %v", err), http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"explanation": explanation,
			"query":       query,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "EXPLAIN not supported", http.StatusNotImplemented)
	}
}
