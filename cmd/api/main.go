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
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/ifaisalabid1/notes-upload-website/internal/config"
	"github.com/ifaisalabid1/notes-upload-website/internal/database"
	"github.com/ifaisalabid1/notes-upload-website/internal/handler"
	appMiddleware "github.com/ifaisalabid1/notes-upload-website/internal/middleware"
	"github.com/ifaisalabid1/notes-upload-website/internal/repository"
	"github.com/ifaisalabid1/notes-upload-website/internal/service"
	"github.com/ifaisalabid1/notes-upload-website/internal/storage"
)

func main() {
	// Config ────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	config.SetupLogger(&cfg.App)

	// Database ───────────────────────────────────────────────────────────
	db, err := database.Open(cfg.DB.Path)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Storage ────────────────────────────────────────────────────────────
	storageService, err := storage.NewR2Storage(cfg.R2)
	if err != nil {
		slog.Error("failed to initialise R2 storage", "error", err)
		os.Exit(1)
	}

	// Wire up layers ─────────────────────────────────────────────────────
	subjectRepo := repository.NewSubjectRepository(db)
	noteRepo := repository.NewNoteRepository(db)

	subjectService := service.NewSubjectService(subjectRepo)
	noteService := service.NewNoteService(noteRepo, subjectRepo, storageService, cfg.Worker)

	subjectHandler := handler.NewSubjectHandler(subjectService)
	noteHandler := handler.NewNoteHandler(noteService)
	healthHandler := handler.NewHealthHandler(db)

	// Middleware ──────────────────────────────────────────────────────────
	rateLimiter := appMiddleware.NewRateLimiter(cfg.Rate.RPS, cfg.Rate.Burst)

	// Router ─────────────────────────────────────────────────────────────
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(appMiddleware.RequestLogger())
	r.Use(chiMiddleware.Recoverer)
	r.Use(appMiddleware.SecureHeaders)
	r.Use(appMiddleware.CORS(cfg.Server.FrontendOrigin))
	r.Use(appMiddleware.MaxBodySize)
	r.Use(rateLimiter.Middleware())

	r.Get("/health", healthHandler)

	r.Route("/api/v1", func(r chi.Router) {
		subjectHandler.RegisterRoutes(r)
		noteHandler.RegisterRoutes(r)
	})

	// HTTP Server ────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           r,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReadHeaderTimeout: 5 * time.Second,
	}

	shutdownError := make(chan error, 1)
	go func() {
		slog.Info("server starting", "port", cfg.Server.Port, "env", cfg.App.Env)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			shutdownError <- err
		}
	}()

	// Graceful Shutdown ──────────────────────────────────────────────────
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

	slog.Info("shutting down gracefully, waiting for in-flight requests...")

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped cleanly")
}
