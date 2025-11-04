package notes

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type CreateNoteRequest struct {
	Title   string `json:"title" validate:"required,min=1,max=200"`
	Content string `json:"content" validate:"max=5000"`
}

type UpdateNoteRequest struct {
	Title   *string `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Content *string `json:"content,omitempty" validate:"omitempty,max=5000"`
}
