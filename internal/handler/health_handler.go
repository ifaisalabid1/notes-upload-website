package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type healthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) http.HandlerFunc {
	h := &healthHandler{pool: pool}
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

	// Ping acquires a connection from the pool and checks liveness.
	if err := h.pool.Ping(r.Context()); err != nil {
		checks["database"] = "unreachable: " + err.Error()
		overallStatus = "degraded"
	} else {
		// Also surface pool stats — useful for debugging connection exhaustion.
		stat := h.pool.Stat()
		checks["database"] = fmt.Sprintf(
			"ok (total=%d idle=%d acquired=%d)",
			stat.TotalConns(),
			stat.IdleConns(),
			stat.AcquiredConns(),
		)
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
