package domain

import (
	"context"
	"time"
)

type Note struct {
	ID          string    `json:"id"          db:"id"`
	SubjectID   string    `json:"subject_id"  db:"subject_id"`
	Title       string    `json:"title"       db:"title"`
	Description string    `json:"description" db:"description"`
	FileKey     string    `json:"-"           db:"file_key"`
	FileType    string    `json:"file_type"   db:"file_type"`
	FileSize    int64     `json:"file_size"   db:"file_size"`
	ViewerURL   string    `json:"viewer_url"  db:"-"`
	CreatedAt   time.Time `json:"created_at"  db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"  db:"updated_at"`
}

type CreateNoteInput struct {
	SubjectID   string `validate:"required"`
	Title       string `validate:"required,min=2,max=200"`
	Description string `validate:"max=1000"`
}

type NoteRepository interface {
	Create(ctx context.Context, note *Note) error
	GetByID(ctx context.Context, id string) (*Note, error)
	ListBySubject(ctx context.Context, subjectID string) ([]*Note, error)
	Delete(ctx context.Context, id string) error
}
