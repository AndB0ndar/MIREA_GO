package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"app/internal/core"
	"app/internal/platform/database"

	"github.com/jackc/pgx/v5"
)

type NoteRepoPG struct {
	ctx context.Context
}

func NewNoteRepoPG() *NoteRepoPG {
	return &NoteRepoPG{
		ctx: context.Background(),
	}
}

// Создание заметки с prepared statement
func (r *NoteRepoPG) Create(note core.Note) (int64, error) {
	query := `
		INSERT INTO notes (title, content) 
		VALUES ($1, $2) 
		RETURNING id, created_at`

	var id int64
	var createdAt time.Time

	err := database.Pool.QueryRow(r.ctx, query, note.Title, note.Content).
		Scan(&id, &createdAt)

	if err != nil {
		return 0, fmt.Errorf("failed to create note: %w", err)
	}

	return id, nil
}

// Получение по ID
func (r *NoteRepoPG) GetByID(id int64) (*core.Note, error) {
	query := `
		SELECT id, title, content, created_at, updated_at
		FROM notes 
		WHERE id = $1`

	note := &core.Note{}
	err := database.Pool.QueryRow(r.ctx, query, id).
		Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, ErrNoteNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	return note, nil
}

// Получение ВСЕХ заметок (старый метод для совместимости)
func (r *NoteRepoPG) GetAll() ([]*core.Note, error) {
	query := `
		SELECT id, title, content, created_at, updated_at
		FROM notes 
		ORDER BY created_at DESC, id DESC
		LIMIT 1000` // Ограничиваем для безопасности

	rows, err := database.Pool.Query(r.ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes: %w", err)
	}
	defer rows.Close()

	var notes []*core.Note
	for rows.Next() {
		note := &core.Note{}
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, note)
	}

	return notes, nil
}

// Пагинация с OFFSET/LIMIT (неоптимизированная)
func (r *NoteRepoPG) GetAllPaginated(limit, offset int) ([]*core.Note, int, error) {
	// Счетчик общего количества
	var total int
	countQuery := `SELECT COUNT(*) FROM notes`
	err := database.Pool.QueryRow(r.ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count notes: %w", err)
	}

	// Получение данных с пагинацией
	query := `
		SELECT id, title, content, created_at, updated_at
		FROM notes 
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2`

	rows, err := database.Pool.Query(r.ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get notes: %w", err)
	}
	defer rows.Close()

	var notes []*core.Note
	for rows.Next() {
		note := &core.Note{}
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, note)
	}

	return notes, total, nil
}

// Keyset-пагинация (оптимизированная)
func (r *NoteRepoPG) GetPaginated(lastID int64, limit int) ([]*core.Note, error) {
	var query string
	var rows pgx.Rows
	var err error

	if lastID == 0 {
		// Первая страница
		query = `
			SELECT id, title, content, created_at, updated_at
			FROM notes 
			ORDER BY created_at DESC, id DESC
			LIMIT $1`
		rows, err = database.Pool.Query(r.ctx, query, limit)
	} else {
		// Последующие страницы
		query = `
			SELECT id, title, content, created_at, updated_at
			FROM notes 
			WHERE (created_at, id) < (
				SELECT created_at, id FROM notes WHERE id = $1
			)
			ORDER BY created_at DESC, id DESC
			LIMIT $2`
		rows, err = database.Pool.Query(r.ctx, query, lastID, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get paginated notes: %w", err)
	}
	defer rows.Close()

	var notes []*core.Note
	for rows.Next() {
		note := &core.Note{}
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, note)
	}

	return notes, nil
}

// Батчинг для устранения N+1
func (r *NoteRepoPG) GetBatch(ids []int64) ([]*core.Note, error) {
	if len(ids) == 0 {
		return []*core.Note{}, nil
	}

	query := `
		SELECT id, title, content, created_at, updated_at
		FROM notes 
		WHERE id = ANY($1)
		ORDER BY id`

	rows, err := database.Pool.Query(r.ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch notes: %w", err)
	}
	defer rows.Close()

	var notes []*core.Note
	for rows.Next() {
		note := &core.Note{}
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, note)
	}

	return notes, nil
}

// Поиск с использованием GIN индекса
func (r *NoteRepoPG) SearchByTitle(query string, limit int) ([]*core.Note, error) {
	sql := `
		SELECT id, title, content, created_at, updated_at
		FROM notes 
		WHERE to_tsvector('simple', title) @@ plainto_tsquery('simple', $1)
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := database.Pool.Query(r.ctx, sql, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search notes: %w", err)
	}
	defer rows.Close()

	var notes []*core.Note
	for rows.Next() {
		note := &core.Note{}
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, note)
	}

	return notes, nil
}

// Обновление заметки
func (r *NoteRepoPG) Update(id int64, updated core.Note) error {
	query := `
		UPDATE notes 
		SET title = COALESCE(NULLIF($1, ''), title),
			content = COALESCE(NULLIF($2, ''), content),
			updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at`

	var updatedAt *time.Time
	err := database.Pool.QueryRow(r.ctx, query, updated.Title, updated.Content, id).
		Scan(&updatedAt)

	if err == pgx.ErrNoRows {
		return ErrNoteNotFound
	}

	return err
}

// Удаление заметки
func (r *NoteRepoPG) Delete(id int64) error {
	query := `DELETE FROM notes WHERE id = $1`

	result, err := database.Pool.Exec(r.ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNoteNotFound
	}

	return nil
}

// EXPLAIN ANALYZE для диагностики
func (r *NoteRepoPG) ExplainQuery(query string, params ...interface{}) (string, error) {
	explainQuery := "EXPLAIN (ANALYZE, BUFFERS, VERBOSE) " + query

	rows, err := database.Pool.Query(r.ctx, explainQuery, params...)
	if err != nil {
		return "", fmt.Errorf("failed to explain query: %w", err)
	}
	defer rows.Close()

	var explanation strings.Builder
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return "", fmt.Errorf("failed to scan explain result: %w", err)
		}
		explanation.WriteString(line + "\n")
	}

	return explanation.String(), nil
}
