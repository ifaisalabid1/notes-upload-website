package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type healthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) http.HandlerFunc {
	h := &healthHandler{db: db}
	return h.handle
}

type healthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

func (h *healthHandler) handle(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)
	overallStatus := "ok"

	ctx := r.Context()
	if err := h.db.PingContext(ctx); err != nil {
		checks["database"] = "unreachable: " + err.Error()
		overallStatus = "degraded"
	} else {
		checks["database"] = "ok"
	}

	status := http.StatusOK
	if overallStatus != "ok" {
		status = http.StatusServiceUnavailable
	}

	resp := healthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC(),
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
