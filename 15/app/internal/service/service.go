package service

import (
	"fmt"

	"app/internal/core"
)

// Service - основной сервис приложения
type Service struct {
	noteRepo NoteRepo
	userRepo UserRepo
}

// NewService создает новый экземпляр сервиса
func NewService(noteRepo NoteRepo, userRepo UserRepo) *Service {
	return &Service{
		noteRepo: noteRepo,
		userRepo: userRepo,
	}
}

// FindIDByEmail возвращает ID пользователя по email
func (s *Service) FindIDByEmail(email string) (int64, error) {
	user, err := s.userRepo.ByEmail(email)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

// GetNoteWithAuthor возвращает заметку с информацией об авторе
func (s *Service) GetNoteWithAuthor(noteID, authorID int64) (*core.Note, *User, error) {
	note, err := s.noteRepo.GetByID(noteID)
	if err != nil {
		return nil, nil, err
	}

	user, err := s.userRepo.ByID(authorID)
	if err != nil {
		return note, nil, nil // Возвращаем заметку даже без автора
	}

	return note, &user, nil
}

// CreateNoteWithValidation создает заметку с валидацией
func (s *Service) CreateNoteWithValidation(title, content string, authorID int64) (*core.Note, error) {
	// Валидация
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if len(title) > 100 {
		return nil, fmt.Errorf("title too long")
	}

	// Проверка существования автора
	_, err := s.userRepo.ByID(authorID)
	if err != nil {
		return nil, fmt.Errorf("author not found: %w", err)
	}

	note := core.Note{
		Title:   title,
		Content: content,
	}

	id, err := s.noteRepo.Create(note)
	if err != nil {
		return nil, err
	}

	return s.noteRepo.GetByID(id)
}
