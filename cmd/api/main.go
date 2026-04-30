package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ifaisalabid1/notes-upload-website/internal/config"
	"github.com/ifaisalabid1/notes-upload-website/internal/database"
	"github.com/ifaisalabid1/notes-upload-website/internal/handler"
	"github.com/ifaisalabid1/notes-upload-website/internal/repository"
	"github.com/ifaisalabid1/notes-upload-website/internal/service"
)

func main() {
	// ── 1. Config ────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	config.SetupLogger(&cfg.App)

	// ── 2. Database ───────────────────────────────────────────────────────────
	db, err := database.Open(cfg.DB.Path)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// ── 3. Wire up layers ─────────────────────────────────────────────────────
	subjectRepo := repository.NewSubjectRepository(db)
	subjectService := service.NewSubjectService(subjectRepo)
	subjectHandler := handler.NewSubjectHandler(subjectService)

	// ── 3. Router ─────────────────────────────────────────────────────────────
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		subjectHandler.RegisterRoutes(r)
	})

	// ── 4. HTTP Server ────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	shutdownError := make(chan error, 1)
	go func() {
		slog.Info("server starting", "port", cfg.Server.Port, "env", cfg.App.Env)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			shutdownError <- err
		}
	}()

	// ── 5. Graceful Shutdown ──────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-shutdownError:
		slog.Error("server error", "error", err)
		os.Exit(1)
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped cleanly")
}
