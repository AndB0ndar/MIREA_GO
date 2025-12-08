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

type SuccessResponse struct {
	Notes  []*Note `json:"notes"`
	Total  int     `json:"total"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
}

type ErrorResponse struct {
	Error   string `json:"error" example:"Error message"`
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"Detailed error message"`
}

type ExplainResponse struct {
	Explanation string `json:"explanation"`
	Query       string `json:"query"`
}

type BatchRequest struct {
	IDs []int64 `json:"ids" example:"1,2,3,4,5"`
}
