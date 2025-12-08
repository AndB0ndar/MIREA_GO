package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"app/internal/core"
	"app/internal/repo"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNoteRepository - мок репозитория для тестов, соответствующий интерфейсу NoteRepository
type mockNoteRepository struct {
	notes  map[int64]*core.Note
	nextID int64
}

func newMockNoteRepository() *mockNoteRepository {
	return &mockNoteRepository{
		notes:  make(map[int64]*core.Note),
		nextID: 1,
	}
}

func (m *mockNoteRepository) Create(note core.Note) (int64, error) {
	note.ID = m.nextID
	note.CreatedAt = time.Now()
	m.notes[m.nextID] = &note
	m.nextID++
	return note.ID, nil
}

func (m *mockNoteRepository) GetByID(id int64) (*core.Note, error) {
	note, exists := m.notes[id]
	if !exists {
		return nil, repo.ErrNoteNotFound
	}
	return note, nil
}

// GetAll - возвращает все заметки (без пагинации)
func (m *mockNoteRepository) GetAll() ([]*core.Note, error) {
	allNotes := make([]*core.Note, 0, len(m.notes))
	for _, note := range m.notes {
		allNotes = append(allNotes, note)
	}
	return allNotes, nil
}

// GetAllPaginated - возвращает заметки с OFFSET/LIMIT пагинацией
func (m *mockNoteRepository) GetAllPaginated(limit, offset int) ([]*core.Note, int, error) {
	// Получаем все заметки
	allNotes := make([]*core.Note, 0, len(m.notes))
	for _, note := range m.notes {
		allNotes = append(allNotes, note)
	}

	// Сортируем по ID для предсказуемости
	for i := 0; i < len(allNotes); i++ {
		for j := i + 1; j < len(allNotes); j++ {
			if allNotes[i].ID > allNotes[j].ID {
				allNotes[i], allNotes[j] = allNotes[j], allNotes[i]
			}
		}
	}

	total := len(allNotes)

	// Применяем пагинацию
	if offset >= len(allNotes) {
		return []*core.Note{}, total, nil
	}

	end := offset + limit
	if end > len(allNotes) {
		end = len(allNotes)
	}

	return allNotes[offset:end], total, nil
}

// GetPaginated - keyset-пагинация
func (m *mockNoteRepository) GetPaginated(lastID int64, limit int) ([]*core.Note, error) {
	var result []*core.Note

	for _, note := range m.notes {
		if lastID == 0 || note.ID < lastID {
			result = append(result, note)
		}
	}

	// Сортируем по ID в порядке убывания
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].ID < result[j].ID {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	// Применяем лимит
	if len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}

// GetBatch - получение заметок по списку ID
func (m *mockNoteRepository) GetBatch(ids []int64) ([]*core.Note, error) {
	var result []*core.Note

	for _, id := range ids {
		if note, exists := m.notes[id]; exists {
			result = append(result, note)
		}
	}

	return result, nil
}

// SearchByTitle - поиск по заголовку
func (m *mockNoteRepository) SearchByTitle(query string, limit int) ([]*core.Note, error) {
	var result []*core.Note

	for _, note := range m.notes {
		if strings.Contains(strings.ToLower(note.Title), strings.ToLower(query)) {
			result = append(result, note)
		}
	}

	// Применяем лимит
	if len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}

func (m *mockNoteRepository) Update(id int64, updated core.Note) error {
	note, exists := m.notes[id]
	if !exists {
		return repo.ErrNoteNotFound
	}

	if updated.Title != "" {
		note.Title = updated.Title
	}
	if updated.Content != "" {
		note.Content = updated.Content
	}
	now := time.Now()
	note.UpdatedAt = &now

	return nil
}

func (m *mockNoteRepository) Delete(id int64) error {
	if _, exists := m.notes[id]; !exists {
		return repo.ErrNoteNotFound
	}
	delete(m.notes, id)
	return nil
}

// ExplainQuery - заглушка для EXPLAIN ANALYZE
func (m *mockNoteRepository) ExplainQuery(query string, params ...interface{}) (string, error) {
	// В тестах просто возвращаем заглушку
	return fmt.Sprintf("EXPLAIN ANALYZE for query: %s", query), nil
}

// TestCreateNote - тесты для CreateNote
func TestCreateNote(t *testing.T) {
	repo := newMockNoteRepository()
	handler := &Handler{Repo: repo}

	tests := []struct {
		name       string
		request    core.NoteCreate
		wantStatus int
		wantError  bool
	}{
		{
			name: "Успешное создание заметки",
			request: core.NoteCreate{
				Title:   "Test Title",
				Content: "Test Content",
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name: "Пустой заголовок",
			request: core.NoteCreate{
				Title:   "",
				Content: "Test Content",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name:       "Пустое тело запроса",
			request:    core.NoteCreate{},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/notes", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.CreateNote(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantError {
				var note core.Note
				err := json.Unmarshal(w.Body.Bytes(), &note)
				require.NoError(t, err)
				assert.Equal(t, tt.request.Title, note.Title)
				assert.Equal(t, tt.request.Content, note.Content)
				assert.NotZero(t, note.ID)
			}
		})
	}
}

// TestGetNote - тесты для GetNote
func TestGetNote(t *testing.T) {
	repo := newMockNoteRepository()
	handler := &Handler{Repo: repo}

	// Создаем тестовую заметку
	note := core.Note{
		Title:   "Test Note",
		Content: "Test Content",
	}
	repo.Create(note)

	tests := []struct {
		name       string
		noteID     string
		wantStatus int
		wantError  bool
	}{
		{
			name:       "Существующая заметка",
			noteID:     "1",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "Несуществующая заметка",
			noteID:     "999",
			wantStatus: http.StatusNotFound,
			wantError:  true,
		},
		{
			name:       "Неверный ID",
			noteID:     "invalid",
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/notes/"+tt.noteID, nil)

			// Создаем контекст с параметром
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.noteID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.GetNote(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantError && tt.noteID == "1" {
				var note core.Note
				err := json.Unmarshal(w.Body.Bytes(), &note)
				require.NoError(t, err)
				assert.Equal(t, "Test Note", note.Title)
			}
		})
	}
}

// TestUpdateNote - тесты для UpdateNote
func TestUpdateNote(t *testing.T) {
	repo := newMockNoteRepository()
	handler := &Handler{Repo: repo}

	// Создаем тестовую заметку
	note := core.Note{
		Title:   "Original Title",
		Content: "Original Content",
	}
	repo.Create(note)

	tests := []struct {
		name       string
		noteID     string
		request    core.NoteUpdate
		wantStatus int
	}{
		{
			name:   "Обновление заголовка",
			noteID: "1",
			request: core.NoteUpdate{
				Title: stringPtr("Updated Title"),
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "Обновление содержимого",
			noteID: "1",
			request: core.NoteUpdate{
				Content: stringPtr("Updated Content"),
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "Обновление обоих полей",
			noteID: "1",
			request: core.NoteUpdate{
				Title:   stringPtr("New Title"),
				Content: stringPtr("New Content"),
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Несуществующая заметка",
			noteID:     "999",
			request:    core.NoteUpdate{},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("PATCH", "/notes/"+tt.noteID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Создаем контекст с параметром
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.noteID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.UpdateNote(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var note core.Note
				err := json.Unmarshal(w.Body.Bytes(), &note)
				require.NoError(t, err)

				if tt.request.Title != nil {
					assert.Equal(t, *tt.request.Title, note.Title)
				}
				if tt.request.Content != nil {
					assert.Equal(t, *tt.request.Content, note.Content)
				}
			}
		})
	}
}

// TestDeleteNote - тесты для DeleteNote
func TestDeleteNote(t *testing.T) {
	repo := newMockNoteRepository()
	handler := &Handler{Repo: repo}

	// Создаем тестовую заметку
	note := core.Note{
		Title:   "Test Note",
		Content: "Test Content",
	}
	repo.Create(note)

	tests := []struct {
		name       string
		noteID     string
		wantStatus int
	}{
		{
			name:       "Удаление существующей заметки",
			noteID:     "1",
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "Удаление несуществующей заметки",
			noteID:     "999",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "Неверный ID",
			noteID:     "invalid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/notes/"+tt.noteID, nil)

			// Создаем контекст с параметром
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.noteID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.DeleteNote(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// TestGetAllNotes - тесты для GetAllNotes
func TestGetAllNotes(t *testing.T) {
	repo := newMockNoteRepository()
	handler := &Handler{Repo: repo}

	// Создаем несколько тестовых заметок
	notes := []core.Note{
		{Title: "Note 1", Content: "Content 1"},
		{Title: "Note 2", Content: "Content 2"},
		{Title: "Note 3", Content: "Content 3"},
	}

	for _, note := range notes {
		repo.Create(note)
	}

	t.Run("Получение всех заметок", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/notes", nil)
		w := httptest.NewRecorder()

		handler.GetAllNotes(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var result []*core.Note
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Len(t, result, 3)
	})

	t.Run("Получение пустого списка", func(t *testing.T) {
		emptyRepo := newMockNoteRepository()
		emptyHandler := &Handler{Repo: emptyRepo}

		req := httptest.NewRequest("GET", "/notes", nil)
		w := httptest.NewRecorder()

		emptyHandler.GetAllNotes(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var result []*core.Note
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Empty(t, result)
	})
}

// Вспомогательная функция для создания указателя на строку
func stringPtr(s string) *string {
	return &s
}
