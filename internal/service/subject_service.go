package service

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/ifaisalabid1/notes-upload-website/internal/domain"
	"github.com/oklog/ulid/v2"
)

type SubjectService struct {
	repo domain.SubjectRepository
}

func NewSubjectService(repo domain.SubjectRepository) *SubjectService {
	return &SubjectService{repo: repo}
}

func (s *SubjectService) Create(ctx context.Context, input domain.CreateSubjectInput) (*domain.Subject, error) {
	slug := slugify(input.Name)

	if _, err := s.repo.GetBySlug(ctx, slug); err == nil {
		slug = fmt.Sprintf("%s-%s", slug, randomSuffix())
	}

	now := time.Now().UTC()
	subject := &domain.Subject{
		ID:          newID(),
		Name:        input.Name,
		Slug:        slug,
		Description: input.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, subject); err != nil {
		return nil, err
	}

	slog.Info("subject created", "id", subject.ID, "name", subject.Name)
	return subject, nil
}

func (s *SubjectService) GetByID(ctx context.Context, id string) (*domain.Subject, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SubjectService) GetBySlug(ctx context.Context, slug string) (*domain.Subject, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *SubjectService) List(ctx context.Context) ([]domain.Subject, error) {
	return s.repo.List(ctx)
}

func (s *SubjectService) Update(ctx context.Context, id string, input domain.UpdateSubjectInput) (*domain.Subject, error) {
	subject, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		subject.Name = *input.Name
		subject.Slug = slugify(*input.Name)
	}
	if input.Description != nil {
		subject.Description = *input.Description
	}

	subject.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, subject); err != nil {
		return nil, err
	}

	slog.Info("subject updated", "id", subject.ID)
	return subject, nil
}

func (s *SubjectService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	slog.Info("subject deleted", "id", id)
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func newID() string {
	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

func randomSuffix() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 4)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
