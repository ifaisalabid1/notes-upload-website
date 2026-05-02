package domain

import (
	"context"
	"time"
)

type Subject struct {
	ID          string    `json:"id"          db:"id"`
	Name        string    `json:"name"        db:"name"`
	Slug        string    `json:"slug"        db:"slug"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at"  db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"  db:"updated_at"`
}

type CreateSubjectInput struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateSubjectInput struct {
	Name        *string `json:"name" validate:"omitzero,min=2,max=100"`
	Description *string `json:"description" validate:"omitzero,max=500"`
}

type SubjectRepository interface {
	Create(ctx context.Context, subject *Subject) error
	GetByID(ctx context.Context, id string) (*Subject, error)
	GetBySlug(ctx context.Context, slug string) (*Subject, error)
	List(ctx context.Context) ([]*Subject, error)
	Update(ctx context.Context, subject *Subject) error
	Delete(ctx context.Context, id string) error
}
