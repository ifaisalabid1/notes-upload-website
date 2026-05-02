package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ifaisalabid1/notes-upload-website/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type noteRepository struct {
	pool *pgxpool.Pool
}

func NewNoteRepository(pool *pgxpool.Pool) domain.NoteRepository {
	return &noteRepository{pool: pool}
}

func (r *noteRepository) Create(ctx context.Context, n *domain.Note) error {
	query := `
		INSERT INTO notes
			(id, subject_id, title, description, file_key, file_type, file_size, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.pool.Exec(ctx, query,
		n.ID, n.SubjectID, n.Title, n.Description,
		n.FileKey, n.FileType, n.FileSize,
		n.CreatedAt, n.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert note: %w", err)
	}
	return nil
}

func (r *noteRepository) GetByID(ctx context.Context, id string) (*domain.Note, error) {
	query := `
		SELECT id, subject_id, title, description, file_key, file_type, file_size, created_at, updated_at
		FROM notes
		WHERE id = $1`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query note by id: %w", err)
	}

	n, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[domain.Note])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("collect note: %w", err)
	}
	return n, nil
}

func (r *noteRepository) ListBySubject(ctx context.Context, subjectID string) ([]*domain.Note, error) {
	query := `
		SELECT id, subject_id, title, description, file_key, file_type, file_size, created_at, updated_at
		FROM notes
		WHERE subject_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, subjectID)
	if err != nil {
		return nil, fmt.Errorf("query notes by subject: %w", err)
	}

	notes, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Note])
	if err != nil {
		return nil, fmt.Errorf("collect notes: %w", err)
	}

	if notes == nil {
		notes = []*domain.Note{}
	}
	return notes, nil
}

func (r *noteRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM notes WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
