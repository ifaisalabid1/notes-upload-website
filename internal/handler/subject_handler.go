package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/ifaisalabid1/notes-upload-website/internal/domain"
	"github.com/ifaisalabid1/notes-upload-website/internal/service"
)

type SubjectHandler struct {
	service  *service.SubjectService
	validate *validator.Validate
}

func NewSubjectHandler(svc *service.SubjectService) *SubjectHandler {
	return &SubjectHandler{
		service:  svc,
		validate: validator.New(),
	}
}

func (h *SubjectHandler) RegisterRoutes(r chi.Router) {
	r.Route("/subjects", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

func (h *SubjectHandler) List(w http.ResponseWriter, r *http.Request) {
	subjects, err := h.service.List(r.Context())
	if err != nil {
		slog.Error("list subjects", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list subjects")
		return
	}
	writeJSON(w, http.StatusOK, subjects)
}

func (h *SubjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateSubjectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		writeError(w, http.StatusUnprocessableEntity, formatValidationErrors(err))
		return
	}

	subject, err := h.service.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			writeError(w, http.StatusConflict, "subject already exists")
			return
		}
		slog.Error("create subject", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create subject")
		return
	}

	writeJSON(w, http.StatusCreated, subject)
}

func (h *SubjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	subject, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "subject not found")
			return
		}
		slog.Error("get subject", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get subject")
		return
	}

	writeJSON(w, http.StatusOK, subject)
}

func (h *SubjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input domain.UpdateSubjectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		writeError(w, http.StatusUnprocessableEntity, formatValidationErrors(err))
		return
	}

	subject, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "subject not found")
			return
		}
		if errors.Is(err, domain.ErrAlreadyExists) {
			writeError(w, http.StatusConflict, "subject name already taken")
			return
		}
		slog.Error("update subject", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to update subject")
		return
	}

	writeJSON(w, http.StatusOK, subject)
}

func (h *SubjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "subject not found")
			return
		}
		slog.Error("delete subject", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to delete subject")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func formatValidationErrors(err error) string {
	if ve, ok := errors.AsType[validator.ValidationErrors](err); ok {
		msgs := make([]string, 0, len(ve))
		for _, e := range ve {
			msgs = append(msgs, e.Field()+": "+e.Tag())
		}
		return strings.Join(msgs, ", ")
	}
	return err.Error()
}
