package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ifaisalabid1/notes-upload-website/internal/domain"
)

type noteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) domain.NoteRepository {
	return &noteRepository{db: db}
}

func (r *noteRepository) Create(ctx context.Context, n *domain.Note) error {
	query := `
		INSERT INTO notes
			(id, subject_id, title, description, file_key, file_type, file_size, created_at, updated_at)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
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
		WHERE id = ?`

	n := &domain.Note{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&n.ID, &n.SubjectID, &n.Title, &n.Description,
		&n.FileKey, &n.FileType, &n.FileSize,
		&n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get note by id: %w", err)
	}
	return n, nil
}

func (r *noteRepository) ListBySubject(ctx context.Context, subjectID string) ([]domain.Note, error) {
	query := `
		SELECT id, subject_id, title, description, file_key, file_type, file_size, created_at, updated_at
		FROM notes
		WHERE subject_id = ?
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, subjectID)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	var notes []domain.Note
	for rows.Next() {
		var n domain.Note
		if err := rows.Scan(
			&n.ID, &n.SubjectID, &n.Title, &n.Description,
			&n.FileKey, &n.FileType, &n.FileSize,
			&n.CreatedAt, &n.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notes: %w", err)
	}

	if notes == nil {
		notes = []domain.Note{}
	}
	return notes, nil
}

func (r *noteRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM notes WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}
