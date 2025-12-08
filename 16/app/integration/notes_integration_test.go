package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"app/internal/core"
	"app/internal/db"
	httpx "app/internal/http"
	"app/internal/http/handlers"
	"app/internal/repo"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDB     *sql.DB
	testRouter *chi.Mux
	testServer *httptest.Server
)

// TestMain настраивает тестовое окружение
func TestMain(m *testing.M) {
	// Получаем DSN из переменных окружения или используем дефолтный
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://test:test@localhost:54321/notes_test?sslmode=disable"
	}

	var err error
	testDB, err = sql.Open("pgx", dsn)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}
	defer testDB.Close()

	// Проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := testDB.PingContext(ctx); err != nil {
		panic(fmt.Sprintf("Failed to ping test database: %v", err))
	}

	// Инициализируем схему
	if err := db.InitSchema(testDB); err != nil {
		panic(fmt.Sprintf("Failed to initialize schema: %v", err))
	}

	// Инициализируем репозиторий и хендлеры
	noteRepo := newTestNoteRepo(testDB)
	handler := &handlers.Handler{Repo: noteRepo}
	testRouter = httpx.NewRouter(handler)

	// Запускаем тесты
	code := m.Run()

	// Очищаем таблицы после тестов
	db.TruncateTables(testDB)

	os.Exit(code)
}

// newTestNoteRepo создает тестовый репозиторий
func newTestNoteRepo(db *sql.DB) repo.NoteRepository {
	return &testNoteRepository{db: db}
}

// testNoteRepository - реализация репозитория для тестов
type testNoteRepository struct {
	db *sql.DB
}

func (r *testNoteRepository) Create(note core.Note) (int64, error) {
	var id int64
	err := r.db.QueryRow(
		"INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING id",
		note.Title, note.Content,
	).Scan(&id)
	return id, err
}

func (r *testNoteRepository) GetByID(id int64) (*core.Note, error) {
	note := &core.Note{}
	err := r.db.QueryRow(
		"SELECT id, title, content, created_at, updated_at FROM notes WHERE id = $1",
		id,
	).Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, repo.ErrNoteNotFound
	}
	return note, err
}

func (r *testNoteRepository) GetAll() ([]*core.Note, error) {
	rows, err := r.db.Query(
		"SELECT id, title, content, created_at, updated_at FROM notes ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*core.Note
	for rows.Next() {
		note := &core.Note{}
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	return notes, nil
}

func (r *testNoteRepository) GetAllPaginated(limit, offset int) ([]*core.Note, int, error) {
	// Реализация для тестов
	allNotes, err := r.GetAll()
	if err != nil {
		return nil, 0, err
	}

	total := len(allNotes)
	if offset >= total {
		return []*core.Note{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return allNotes[offset:end], total, nil
}

func (r *testNoteRepository) GetPaginated(lastID int64, limit int) ([]*core.Note, error) {
	// Реализация для тестов
	return nil, nil
}

func (r *testNoteRepository) GetBatch(ids []int64) ([]*core.Note, error) {
	// Реализация для тестов
	return nil, nil
}

func (r *testNoteRepository) SearchByTitle(query string, limit int) ([]*core.Note, error) {
	// Реализация для тестов
	return nil, nil
}

func (r *testNoteRepository) Update(id int64, updated core.Note) error {
	result, err := r.db.Exec(
		"UPDATE notes SET title = $1, content = $2 WHERE id = $3",
		updated.Title, updated.Content, id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return repo.ErrNoteNotFound
	}

	return nil
}

func (r *testNoteRepository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM notes WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return repo.ErrNoteNotFound
	}

	return nil
}

func (r *testNoteRepository) ExplainQuery(query string, params ...interface{}) (string, error) {
	// Реализация для тестов
	return "EXPLAIN output for test", nil
}

// setupTest подготавливает тестовое окружение
func setupTest(t *testing.T) {
	t.Helper()
	// Очищаем таблицы перед каждым тестом
	require.NoError(t, db.TruncateTables(testDB))

	// Создаем тестовый сервер
	testServer = httptest.NewServer(testRouter)
	t.Cleanup(func() {
		testServer.Close()
	})
}

// TestCreateNoteIntegration тестирует создание заметки
func TestCreateNoteIntegration(t *testing.T) {
	setupTest(t)

	// Подготавливаем запрос
	requestBody := `{"title": "Integration Test Note", "content": "This is integration test content"}`
	req, err := http.NewRequest("POST", testServer.URL+"/notes", bytes.NewBufferString(requestBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Проверяем статус
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Проверяем заголовки
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	// Проверяем тело ответа
	var responseNote core.Note
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &responseNote))

	assert.NotZero(t, responseNote.ID)
	assert.Equal(t, "Integration Test Note", responseNote.Title)
	assert.Equal(t, "This is integration test content", responseNote.Content)
	assert.NotZero(t, responseNote.CreatedAt)

	// Проверяем, что заметка действительно сохранена в БД
	var dbNote core.Note
	err = testDB.QueryRow(
		"SELECT id, title, content, created_at FROM notes WHERE id = $1",
		responseNote.ID,
	).Scan(&dbNote.ID, &dbNote.Title, &dbNote.Content, &dbNote.CreatedAt)

	require.NoError(t, err)
	assert.Equal(t, responseNote.ID, dbNote.ID)
	assert.Equal(t, responseNote.Title, dbNote.Title)
	assert.Equal(t, responseNote.Content, dbNote.Content)
}

// TestGetNoteIntegration тестирует получение заметки
func TestGetNoteIntegration(t *testing.T) {
	setupTest(t)

	// Сначала создаем заметку
	var noteID int64
	err := testDB.QueryRow(
		"INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING id",
		"Test Note for GET", "Test content for GET",
	).Scan(&noteID)
	require.NoError(t, err)

	// Получаем заметку через API
	resp, err := http.Get(fmt.Sprintf("%s/notes/%d", testServer.URL, noteID))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Проверяем статус
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Проверяем тело ответа
	var responseNote core.Note
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &responseNote))

	assert.Equal(t, noteID, responseNote.ID)
	assert.Equal(t, "Test Note for GET", responseNote.Title)
	assert.Equal(t, "Test content for GET", responseNote.Content)
}

// TestGetNoteNotFoundIntegration тестирует получение несуществующей заметки
func TestGetNoteNotFoundIntegration(t *testing.T) {
	setupTest(t)

	// Пытаемся получить несуществующую заметку
	resp, err := http.Get(fmt.Sprintf("%s/notes/%d", testServer.URL, 99999))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Проверяем статус
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestUpdateNoteIntegration тестирует обновление заметки
func TestUpdateNoteIntegration(t *testing.T) {
	setupTest(t)

	// Сначала создаем заметку
	var noteID int64
	err := testDB.QueryRow(
		"INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING id",
		"Original Title", "Original Content",
	).Scan(&noteID)
	require.NoError(t, err)

	// Обновляем заметку
	updateBody := `{"title": "Updated Title", "content": "Updated Content"}`
	req, err := http.NewRequest("PATCH",
		fmt.Sprintf("%s/notes/%d", testServer.URL, noteID),
		bytes.NewBufferString(updateBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Проверяем статус
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Проверяем, что заметка обновлена в БД
	var dbTitle, dbContent string
	var dbUpdatedAt *time.Time
	err = testDB.QueryRow(
		"SELECT title, content, updated_at FROM notes WHERE id = $1",
		noteID,
	).Scan(&dbTitle, &dbContent, &dbUpdatedAt)

	require.NoError(t, err)
	assert.Equal(t, "Updated Title", dbTitle)
	assert.Equal(t, "Updated Content", dbContent)
	assert.NotNil(t, dbUpdatedAt)
}

// TestDeleteNoteIntegration тестирует удаление заметки
func TestDeleteNoteIntegration(t *testing.T) {
	setupTest(t)

	// Сначала создаем заметку
	var noteID int64
	err := testDB.QueryRow(
		"INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING id",
		"Note to Delete", "Content to delete",
	).Scan(&noteID)
	require.NoError(t, err)

	// Удаляем заметку
	req, err := http.NewRequest("DELETE",
		fmt.Sprintf("%s/notes/%d", testServer.URL, noteID), nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Проверяем статус
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Проверяем, что заметка удалена из БД
	var count int
	err = testDB.QueryRow(
		"SELECT COUNT(*) FROM notes WHERE id = $1",
		noteID,
	).Scan(&count)

	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestGetAllNotesIntegration тестирует получение всех заметок
func TestGetAllNotesIntegration(t *testing.T) {
	setupTest(t)

	// Создаем несколько заметок
	notes := []struct {
		title   string
		content string
	}{
		{"Note 1", "Content 1"},
		{"Note 2", "Content 2"},
		{"Note 3", "Content 3"},
	}

	for _, note := range notes {
		_, err := testDB.Exec(
			"INSERT INTO notes (title, content) VALUES ($1, $2)",
			note.title, note.content,
		)
		require.NoError(t, err)
	}

	// Получаем все заметки
	resp, err := http.Get(testServer.URL + "/notes")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Проверяем статус
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Проверяем тело ответа
	var responseNotes []core.Note
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &responseNotes))

	assert.Len(t, responseNotes, 3)

	// Проверяем, что все заметки присутствуют
	titles := make(map[string]bool)
	for _, note := range responseNotes {
		titles[note.Title] = true
	}

	assert.True(t, titles["Note 1"])
	assert.True(t, titles["Note 2"])
	assert.True(t, titles["Note 3"])
}

// TestCreateNoteValidationIntegration тестирует валидацию при создании заметки
func TestCreateNoteValidationIntegration(t *testing.T) {
	setupTest(t)

	// Пытаемся создать заметку без заголовка
	requestBody := `{"content": "Content without title"}`
	req, err := http.NewRequest("POST", testServer.URL+"/notes", bytes.NewBufferString(requestBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Проверяем статус
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Проверяем, что в БД не добавилась запись
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM notes").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestConcurrentRequestsIntegration тестирует конкурентные запросы
func TestConcurrentRequestsIntegration(t *testing.T) {
	setupTest(t)

	const numRequests = 10
	errors := make(chan error, numRequests)

	// Запускаем несколько конкурентных запросов
	for i := 0; i < numRequests; i++ {
		go func(id int) {
			requestBody := fmt.Sprintf(`{"title": "Concurrent Note %d", "content": "Content %d"}`, id, id)
			req, err := http.NewRequest("POST", testServer.URL+"/notes", bytes.NewBufferString(requestBody))
			if err != nil {
				errors <- err
				return
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				errors <- fmt.Errorf("unexpected status: %d", resp.StatusCode)
				return
			}

			errors <- nil
		}(i)
	}

	// Ожидаем завершения всех горутин
	for i := 0; i < numRequests; i++ {
		err := <-errors
		assert.NoError(t, err)
	}

	// Проверяем, что все заметки созданы
	var count int
	err := testDB.QueryRow("SELECT COUNT(*) FROM notes").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, numRequests, count)
}
