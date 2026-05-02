package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ifaisalabid1/notes-upload-website/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type subjectRepository struct {
	pool *pgxpool.Pool
}

func NewSubjectRepository(pool *pgxpool.Pool) domain.SubjectRepository {
	return &subjectRepository{pool: pool}
}

func (r *subjectRepository) Create(ctx context.Context, s *domain.Subject) error {
	query := `
		INSERT INTO subjects (id, name, slug, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.pool.Exec(ctx, query,
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
		FROM subjects
		WHERE id = $1`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query subject by id: %w", err)
	}

	s, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[domain.Subject])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("collect subject: %w", err)
	}
	return s, nil
}

func (r *subjectRepository) GetBySlug(ctx context.Context, slug string) (*domain.Subject, error) {
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM subjects
		WHERE slug = $1`

	rows, err := r.pool.Query(ctx, query, slug)
	if err != nil {
		return nil, fmt.Errorf("query subject by slug: %w", err)
	}

	s, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[domain.Subject])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("collect subject: %w", err)
	}
	return s, nil
}

func (r *subjectRepository) List(ctx context.Context) ([]*domain.Subject, error) {
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM subjects
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query subjects: %w", err)
	}

	subjects, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Subject])
	if err != nil {
		return nil, fmt.Errorf("collect subjects: %w", err)
	}

	// Always return an empty slice, never nil — cleaner JSON responses.
	if subjects == nil {
		subjects = []*domain.Subject{}
	}
	return subjects, nil
}

func (r *subjectRepository) Update(ctx context.Context, s *domain.Subject) error {
	query := `
		UPDATE subjects
		SET name = $1, slug = $2, description = $3, updated_at = $4
		WHERE id = $5`

	tag, err := r.pool.Exec(ctx, query,
		s.Name, s.Slug, s.Description, s.UpdatedAt, s.ID,
	)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return fmt.Errorf("subject with this name or slug: %w", domain.ErrAlreadyExists)
		}
		return fmt.Errorf("update subject: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *subjectRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM subjects WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete subject: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// isUniqueConstraintErr detects PostgreSQL unique_violation (code 23505).
func isUniqueConstraintErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
