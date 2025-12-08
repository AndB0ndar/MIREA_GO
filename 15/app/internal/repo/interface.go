package repo

import (
	"errors"

	"app/internal/core"
)

var (
	ErrNoteNotFound = errors.New("note not found")
)

type NoteRepository interface {
	Create(note core.Note) (int64, error)
	GetByID(id int64) (*core.Note, error)
	GetAll() ([]*core.Note, error)                                // ВАЖНО: Вернуть старую сигнатуру для совместимости
	GetAllPaginated(limit, offset int) ([]*core.Note, int, error) // Новая версия с пагинацией
	GetPaginated(lastID int64, limit int) ([]*core.Note, error)   // Keyset-пагинация
	GetBatch(ids []int64) ([]*core.Note, error)                   // Батчинг для N+1
	Update(id int64, updated core.Note) error
	Delete(id int64) error
	SearchByTitle(query string, limit int) ([]*core.Note, error)      // Поиск с индексами
	ExplainQuery(query string, params ...interface{}) (string, error) // Для EXPLAIN ANALYZE
}
