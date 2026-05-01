package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	App    AppConfig
	Server ServerConfig
	DB     DBConfig
	R2     R2Config
	Rate   RateLimitConfig
	Worker WorkerConfig
}

type AppConfig struct {
	Env string `env:"APP_ENV" envDefault:"development"`
}

type ServerConfig struct {
	Port         int           `env:"SERVER_PORT"          envDefault:"8080"`
	ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT"  envDefault:"10s"`
	WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout  time.Duration `env:"SERVER_IDLE_TIMEOUT"  envDefault:"60s"`
}

type DBConfig struct {
	Path string `env:"DATABASE_PATH" envDefault:"./data/notes.sqlite"`
}

type R2Config struct {
	AccountID       string `env:"R2_ACCOUNT_ID"        required:"true"`
	AccessKeyID     string `env:"R2_ACCESS_KEY_ID"     required:"true"`
	SecretAccessKey string `env:"R2_SECRET_ACCESS_KEY" required:"true"`
	BucketName      string `env:"R2_BUCKET_NAME"       required:"true"`
}

type WorkerConfig struct {
	BaseURL string `env:"WORKER_BASE_URL" required:"true"`
	Secret  string `env:"WORKER_SECRET"   required:"true"`
}

type RateLimitConfig struct {
	RPS   float64 `env:"RATE_LIMIT_RPS"   envDefault:"10"`
	Burst int     `env:"RATE_LIMIT_BURST" envDefault:"20"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Info("no .env file found, reading config from environment")
	}

	cfg := &Config{}

	if err := env.Parse(&cfg.App); err != nil {
		return nil, err
	}
	if err := env.Parse(&cfg.Server); err != nil {
		return nil, err
	}
	if err := env.Parse(&cfg.DB); err != nil {
		return nil, err
	}
	if err := env.Parse(&cfg.R2); err != nil {
		return nil, err
	}
	if err := env.Parse(&cfg.Rate); err != nil {
		return nil, err
	}
	if err := env.Parse(&cfg.Worker); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (a *AppConfig) IsDevelopment() bool {
	return a.Env == "development"
}

func SetupLogger(cfg *AppConfig) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: !cfg.IsDevelopment(),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}

	if cfg.IsDevelopment() {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}
