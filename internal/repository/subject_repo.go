package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ifaisalabid1/notes-upload-website/internal/domain"
)

type subjectRepository struct {
	db *sql.DB
}

func NewSubjectRepository(db *sql.DB) domain.SubjectRepository {
	return &subjectRepository{db: db}
}

func (r *subjectRepository) Create(ctx context.Context, s *domain.Subject) error {
	query := `
		INSERT INTO subjects (id, name, slug, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.Name, s.Slug, s.Description, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return fmt.Errorf("subject with this name or slug: %w", domain.ErrAlreadyExists)
		}
		return fmt.Errorf("insert subject: %w", err)
	}
	return nil
}

func (r *subjectRepository) GetByID(ctx context.Context, id string) (*domain.Subject, error) {
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM subjects WHERE id = ?`

	s := &domain.Subject{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.Slug, &s.Description, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get subject by id: %w", err)
	}
	return s, nil
}

func (r *subjectRepository) GetBySlug(ctx context.Context, slug string) (*domain.Subject, error) {
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM subjects WHERE slug = ?`

	s := &domain.Subject{}
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&s.ID, &s.Name, &s.Slug, &s.Description, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get subject by slug: %w", err)
	}
	return s, nil
}

func (r *subjectRepository) List(ctx context.Context) ([]domain.Subject, error) {
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM subjects ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	defer rows.Close()

	var subjects []domain.Subject
	for rows.Next() {
		var s domain.Subject
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Slug, &s.Description, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan subject: %w", err)
		}
		subjects = append(subjects, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subjects: %w", err)
	}

	if subjects == nil {
		subjects = []domain.Subject{}
	}

	return subjects, nil
}

func (r *subjectRepository) Update(ctx context.Context, s *domain.Subject) error {
	query := `
		UPDATE subjects
		SET name = ?, slug = ?, description = ?, updated_at = ?
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		s.Name, s.Slug, s.Description, s.UpdatedAt, s.ID,
	)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return fmt.Errorf("subject with this name or slug: %w", domain.ErrAlreadyExists)
		}
		return fmt.Errorf("update subject: %w", err)
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

func (r *subjectRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM subjects WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete subject: %w", err)
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

func isUniqueConstraintErr(err error) bool {
	return err != nil && (containsStr(err.Error(), "UNIQUE constraint failed"))
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
