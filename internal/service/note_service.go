package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/ifaisalabid1/notes-upload-website/internal/config"
	"github.com/ifaisalabid1/notes-upload-website/internal/domain"
	"github.com/ifaisalabid1/notes-upload-website/internal/storage"
)

type NoteService struct {
	noteRepo    domain.NoteRepository
	subjectRepo domain.SubjectRepository
	store       domain.StorageService
	workerCfg   config.WorkerConfig
}

func NewNoteService(
	noteRepo domain.NoteRepository,
	subjectRepo domain.SubjectRepository,
	store domain.StorageService,
	workerCfg config.WorkerConfig,
) *NoteService {
	return &NoteService{
		noteRepo:    noteRepo,
		subjectRepo: subjectRepo,
		store:       store,
		workerCfg:   workerCfg,
	}
}

func (s *NoteService) Upload(ctx context.Context,
	input domain.CreateNoteInput,
	reader io.Reader,
	size int64) (*domain.Note, error) {

	if _, err := s.subjectRepo.GetByID(ctx, input.SubjectID); err != nil {
		return nil, fmt.Errorf("subject lookup: %w", err)
	}

	mimeType, fileType, fullReader, err := storage.DetectFileType(reader)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrInvalidInput, err.Error())
	}

	noteID := newID()
	ext := storage.ExtensionForMIME(mimeType)
	key := storage.NoteKey(input.SubjectID, noteID, ext)

	if err := s.store.Upload(ctx, domain.UploadInput{
		Key:         key,
		Body:        fullReader,
		Size:        size,
		ContentType: mimeType,
		FileType:    fileType,
	}); err != nil {
		return nil, fmt.Errorf("upload to R2: %w", err)
	}

	now := time.Now().UTC()
	note := &domain.Note{
		ID:          noteID,
		SubjectID:   input.SubjectID,
		Title:       input.Title,
		Description: input.Description,
		FileKey:     key,
		FileType:    string(fileType),
		FileSize:    size,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.noteRepo.Create(ctx, note); err != nil {

		if delErr := s.store.Delete(ctx, key); delErr != nil {
			slog.Error("failed to clean up R2 object after DB error",
				"key", key, "error", delErr)
		}
		return nil, fmt.Errorf("persist note: %w", err)
	}

	note.ViewerURL = s.buildViewerURL(key)

	slog.Info("note uploaded", "id", note.ID, "subject_id", input.SubjectID, "size", size)
	return note, nil
}

func (s *NoteService) GetByID(ctx context.Context, id string) (*domain.Note, error) {
	note, err := s.noteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	note.ViewerURL = s.buildViewerURL(note.FileKey)
	return note, nil
}

func (s *NoteService) ListBySubject(ctx context.Context, subjectID string) ([]domain.Note, error) {
	if _, err := s.subjectRepo.GetByID(ctx, subjectID); err != nil {
		return nil, err
	}

	notes, err := s.noteRepo.ListBySubject(ctx, subjectID)
	if err != nil {
		return nil, err
	}

	for _, n := range notes {
		n.ViewerURL = s.buildViewerURL(n.FileKey)
	}
	return notes, nil
}

func (s *NoteService) Delete(ctx context.Context, id string) error {
	note, err := s.noteRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.noteRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete note record: %w", err)
	}

	if err := s.store.Delete(ctx, note.FileKey); err != nil {

		slog.Error("failed to delete R2 object",
			"key", note.FileKey, "note_id", id, "error", err)
	}

	slog.Info("note deleted", "id", id)
	return nil
}

func (s *NoteService) buildViewerURL(key string) string {
	return fmt.Sprintf("%s/view/%s", s.workerCfg.BaseURL, key)
}
