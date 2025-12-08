package service

import (
	"testing"
	"time"

	"app/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubUserRepo - заглушка репозитория пользователей
type stubUserRepo struct {
	users     map[string]User
	usersByID map[int64]User
}

func (r *stubUserRepo) ByEmail(email string) (User, error) {
	user, ok := r.users[email]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (r *stubUserRepo) ByID(id int64) (User, error) {
	user, ok := r.usersByID[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (r *stubUserRepo) Save(user User) (int64, error) {
	// Простая имитация сохранения
	if user.ID == 0 {
		user.ID = int64(len(r.usersByID) + 1)
	}
	r.users[user.Email] = user
	r.usersByID[user.ID] = user
	return user.ID, nil
}

// mockNoteRepo - мок репозитория заметок
type mockNoteRepo struct {
	notes map[int64]*core.Note
}

func (m *mockNoteRepo) GetByID(id int64) (*core.Note, error) {
	note, ok := m.notes[id]
	if !ok {
		return nil, ErrNotFound
	}
	return note, nil
}

func (m *mockNoteRepo) GetAll() ([]*core.Note, error) {
	notes := make([]*core.Note, 0, len(m.notes))
	for _, note := range m.notes {
		notes = append(notes, note)
	}
	return notes, nil
}

func (m *mockNoteRepo) Create(note core.Note) (int64, error) {
	newID := int64(len(m.notes) + 1)
	note.ID = newID
	note.CreatedAt = time.Now()
	m.notes[newID] = &note
	return newID, nil
}

func (m *mockNoteRepo) Update(id int64, updated core.Note) error {
	note, ok := m.notes[id]
	if !ok {
		return ErrNotFound
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

func (m *mockNoteRepo) Delete(id int64) error {
	if _, ok := m.notes[id]; !ok {
		return ErrNotFound
	}
	delete(m.notes, id)
	return nil
}

// TestService_FindIDByEmail - тесты для FindIDByEmail
func TestService_FindIDByEmail(t *testing.T) {
	// Настройка стабов
	userRepo := &stubUserRepo{
		users: map[string]User{
			"test@example.com":  {ID: 1, Email: "test@example.com", Name: "Test User"},
			"admin@example.com": {ID: 2, Email: "admin@example.com", Name: "Admin"},
		},
		usersByID: map[int64]User{
			1: {ID: 1, Email: "test@example.com", Name: "Test User"},
			2: {ID: 2, Email: "admin@example.com", Name: "Admin"},
		},
	}

	noteRepo := &mockNoteRepo{
		notes: make(map[int64]*core.Note),
	}

	service := NewService(noteRepo, userRepo)

	t.Run("Найден существующий пользователь", func(t *testing.T) {
		id, err := service.FindIDByEmail("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, int64(1), id)
	})

	t.Run("Пользователь не найден", func(t *testing.T) {
		id, err := service.FindIDByEmail("nonexistent@example.com")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Equal(t, int64(0), id)
	})

	t.Run("Пустой email", func(t *testing.T) {
		id, err := service.FindIDByEmail("")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Equal(t, int64(0), id)
	})
}

// TestService_CreateNoteWithValidation - тесты для CreateNoteWithValidation
func TestService_CreateNoteWithValidation(t *testing.T) {
	userRepo := &stubUserRepo{
		usersByID: map[int64]User{
			1: {ID: 1, Email: "test@example.com", Name: "Test User"},
		},
		users: map[string]User{},
	}

	noteRepo := &mockNoteRepo{
		notes: make(map[int64]*core.Note),
	}

	service := NewService(noteRepo, userRepo)

	t.Run("Успешное создание заметки", func(t *testing.T) {
		note, err := service.CreateNoteWithValidation("Test Title", "Test Content", 1)
		require.NoError(t, err)
		assert.NotNil(t, note)
		assert.Equal(t, "Test Title", note.Title)
		assert.Equal(t, "Test Content", note.Content)
		assert.NotZero(t, note.ID)
	})

	t.Run("Пустой заголовок", func(t *testing.T) {
		note, err := service.CreateNoteWithValidation("", "Test Content", 1)
		require.Error(t, err)
		assert.Nil(t, note)
		assert.Contains(t, err.Error(), "title is required")
	})

	t.Run("Слишком длинный заголовок", func(t *testing.T) {
		longTitle := "This is a very long title that exceeds the maximum allowed length of 100 characters and should trigger an error"
		note, err := service.CreateNoteWithValidation(longTitle, "Test Content", 1)
		require.Error(t, err)
		assert.Nil(t, note)
		assert.Contains(t, err.Error(), "title too long")
	})

	t.Run("Автор не найден", func(t *testing.T) {
		note, err := service.CreateNoteWithValidation("Test Title", "Test Content", 999)
		require.Error(t, err)
		assert.Nil(t, note)
		assert.Contains(t, err.Error(), "author not found")
	})
}

// TestService_GetNoteWithAuthor - тесты для GetNoteWithAuthor
func TestService_GetNoteWithAuthor(t *testing.T) {
	userRepo := &stubUserRepo{
		usersByID: map[int64]User{
			1: {ID: 1, Email: "author@example.com", Name: "Note Author"},
			2: {ID: 2, Email: "other@example.com", Name: "Other User"},
		},
		users: map[string]User{},
	}

	noteRepo := &mockNoteRepo{
		notes: map[int64]*core.Note{
			1: {
				ID:        1,
				Title:     "Test Note",
				Content:   "Test Content",
				CreatedAt: time.Now(),
			},
		},
	}

	service := NewService(noteRepo, userRepo)

	t.Run("Найдены и заметка и автор", func(t *testing.T) {
		note, author, err := service.GetNoteWithAuthor(1, 1)
		require.NoError(t, err)
		assert.NotNil(t, note)
		assert.NotNil(t, author)
		assert.Equal(t, "Test Note", note.Title)
		assert.Equal(t, "Note Author", author.Name)
	})

	t.Run("Заметка не найдена", func(t *testing.T) {
		note, author, err := service.GetNoteWithAuthor(999, 1)
		require.Error(t, err)
		assert.Nil(t, note)
		assert.Nil(t, author)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("Автор не найден", func(t *testing.T) {
		note, author, err := service.GetNoteWithAuthor(1, 999)
		require.NoError(t, err) // Заметка возвращается даже без автора
		assert.NotNil(t, note)
		assert.Nil(t, author) // Автор nil
		assert.Equal(t, "Test Note", note.Title)
	})
}
