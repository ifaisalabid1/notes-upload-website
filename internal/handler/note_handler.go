package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/ifaisalabid1/notes-upload-website/internal/domain"
	"github.com/ifaisalabid1/notes-upload-website/internal/service"
	"github.com/ifaisalabid1/notes-upload-website/internal/storage"
)

const (
	maxMemory = 10 << 20
)

type NoteHandler struct {
	service  *service.NoteService
	validate *validator.Validate
}

func NewNoteHandler(svc *service.NoteService) *NoteHandler {
	return &NoteHandler{
		service:  svc,
		validate: validator.New(),
	}
}

func (h *NoteHandler) RegisterRoutes(r chi.Router) {

	r.Get("/subjects/{subjectID}/notes", h.ListBySubject)

	r.Route("/notes", func(r chi.Router) {
		r.Post("/", h.Upload)
		r.Get("/{id}", h.GetByID)
		r.Delete("/{id}", h.Delete)
	})
}

func (h *NoteHandler) Upload(w http.ResponseWriter, r *http.Request) {

	r.Body = http.MaxBytesReader(w, r.Body, storage.MaxUploadSize)

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			writeError(w, http.StatusRequestEntityTooLarge,
				"file exceeds the 50 MB limit")
			return
		}
		writeError(w, http.StatusBadRequest, "failed to parse form")
		return
	}

	input := domain.CreateNoteInput{
		SubjectID:   strings.TrimSpace(r.FormValue("subject_id")),
		Title:       strings.TrimSpace(r.FormValue("title")),
		Description: strings.TrimSpace(r.FormValue("description")),
	}

	if err := h.validate.Struct(input); err != nil {
		writeError(w, http.StatusUnprocessableEntity, formatValidationErrors(err))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	note, err := h.service.Upload(r.Context(), input, file, header.Size)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			writeError(w, http.StatusNotFound, "subject not found")
		case errors.Is(err, domain.ErrInvalidInput):
			writeError(w, http.StatusUnsupportedMediaType, err.Error())
		default:
			slog.Error("upload note", "error", err)
			writeError(w, http.StatusInternalServerError, "failed to upload note")
		}
		return
	}

	writeJSON(w, http.StatusCreated, note)
}

func (h *NoteHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	note, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "note not found")
			return
		}
		slog.Error("get note", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get note")
		return
	}

	writeJSON(w, http.StatusOK, note)
}

func (h *NoteHandler) ListBySubject(w http.ResponseWriter, r *http.Request) {
	subjectID := chi.URLParam(r, "subjectID")

	notes, err := h.service.ListBySubject(r.Context(), subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "subject not found")
			return
		}
		slog.Error("list notes", "subject_id", subjectID, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list notes")
		return
	}

	writeJSON(w, http.StatusOK, notes)
}

func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "note not found")
			return
		}
		slog.Error("delete note", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to delete note")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
