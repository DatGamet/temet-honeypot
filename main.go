package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"temet-honeypot/pkg/bot"
	"temet-honeypot/pkg/database"
	"time"

	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
		}),
	))

	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, relying on environment variables")
	}

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		slog.Error("DISCORD_TOKEN is not set")
		os.Exit(1)
	}

	db, err := database.Connect(context.Background())

	b, err := bot.New(token, db)
	if err != nil {
		slog.Error("failed to build bot", "err", err)
		os.Exit(1)
	}

	if err := b.Run(context.Background()); err != nil {
		slog.Error("failed to start bot", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	slog.Info("shutting down…")
	if err := b.Client.Shutdown(); err != nil {
		slog.Error("shutdown error", "err", err)
	}
}
