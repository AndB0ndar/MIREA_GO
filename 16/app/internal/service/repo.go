package service

import (
	"app/internal/core"
	"fmt"
)

type User struct {
	ID    int64
	Email string
	Name  string
}

var ErrNotFound = fmt.Errorf("not found")

type UserRepo interface {
	ByEmail(email string) (User, error)
	ByID(id int64) (User, error)
	Save(user User) (int64, error)
}

type NoteRepo interface {
	GetByID(id int64) (*core.Note, error)
	GetAll() ([]*core.Note, error)
	Create(note core.Note) (int64, error)
	Update(id int64, updated core.Note) error
	Delete(id int64) error
}
