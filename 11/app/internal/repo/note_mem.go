package repo

import (
	"errors"
	"sync"
	"time"

	"app/internal/core"
)

var (
	ErrNoteNotFound = errors.New("note not found")
)

type NoteRepository interface {
	Create(note core.Note) (int64, error)
	GetByID(id int64) (*core.Note, error)
	GetAll() ([]*core.Note, error)
	Update(id int64, updated core.Note) error
	Delete(id int64) error
}

type NoteRepoMem struct {
	mu     sync.RWMutex
	notes  map[int64]*core.Note
	nextID int64
}

func NewNoteRepoMem() *NoteRepoMem {
	return &NoteRepoMem{
		notes:  make(map[int64]*core.Note),
		nextID: 1,
	}
}

func (r *NoteRepoMem) Create(note core.Note) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	note.ID = r.nextID
	note.CreatedAt = time.Now()
	r.notes[note.ID] = &note
	r.nextID++

	return note.ID, nil
}

func (r *NoteRepoMem) GetByID(id int64) (*core.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	note, exists := r.notes[id]
	if !exists {
		return nil, ErrNoteNotFound
	}

	return note, nil
}

func (r *NoteRepoMem) GetAll() ([]*core.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	notes := make([]*core.Note, 0, len(r.notes))
	for _, note := range r.notes {
		notes = append(notes, note)
	}

	return notes, nil
}

func (r *NoteRepoMem) Update(id int64, updated core.Note) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	note, exists := r.notes[id]
	if !exists {
		return ErrNoteNotFound
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

func (r *NoteRepoMem) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.notes[id]; !exists {
		return ErrNoteNotFound
	}

	delete(r.notes, id)
	return nil
}
