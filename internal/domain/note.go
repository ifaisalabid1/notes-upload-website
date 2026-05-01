package domain

import (
	"context"
	"time"
)

type Note struct {
	ID          string    `json:"id"`
	SubjectID   string    `json:"subject_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	FileKey     string    `json:"-"`
	FileType    string    `json:"file_type"`
	FileSize    int64     `json:"file_size"`
	ViewerURL   string    `json:"viewer_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateNoteInput struct {
	SubjectID   string `validate:"required"`
	Title       string `validate:"required,min=2,max=200"`
	Description string `validate:"max=1000"`
}

type NoteRepository interface {
	Create(ctx context.Context, note *Note) error
	GetByID(ctx context.Context, id string) (*Note, error)
	ListBySubject(ctx context.Context, subjectID string) ([]Note, error)
	Delete(ctx context.Context, id string) error
}
