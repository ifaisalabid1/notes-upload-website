package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type envelope struct {
	Data  any    `json:"data,omitzero"`
	Error string `json:"error,omitzero"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(envelope{Data: data}); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(envelope{Error: msg}); err != nil {
		slog.Error("failed to encode error response", "error", err)
	}
}
