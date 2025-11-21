package core

import "time"

type Note struct {
	ID        int64      `json:"id" example:"1"`
	Title     string     `json:"title" example:"Моя заметка"`
	Content   string     `json:"content" example:"Содержание заметки"`
	CreatedAt time.Time  `json:"createdAt" example:"2025-01-15T10:30:00Z"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty" example:"2025-01-15T11:00:00Z"`
}

type NoteCreate struct {
	Title   string `json:"title" example:"Новая заметка"`
	Content string `json:"content" example:"Текст заметки"`
}

type NoteUpdate struct {
	Title   *string `json:"title,omitempty" example:"Обновленный заголовок"`
	Content *string `json:"content,omitempty" example:"Обновленное содержание"`
}
