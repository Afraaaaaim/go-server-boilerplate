package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Afraaaaaim/go-server-boilerplate/internal/config"
	"github.com/Afraaaaaim/go-server-boilerplate/internal/middleware"
	"github.com/Afraaaaaim/go-server-boilerplate/internal/observability"
	"github.com/Afraaaaaim/go-server-boilerplate/internal/server"
)

func main() {
	// Exit code 1 on any startup failure
	if err := run(); err != nil {
		slog.Error("startup failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	// 1. Load config first — everything depends on it
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// 2. Init logger — must happen before any slog calls
	if err := observability.InitLogger(cfg); err != nil {
		return err
	}

	slog.Info("starting up",
		slog.String("log_level", cfg.LogLevel),
		slog.String("log_format", cfg.LogFormat),
	)

	// 3. Init tracer — no-op if OTLP_ENDPOINT is empty
	ctx := context.Background()
	tp, err := observability.InitTracer(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			slog.Error("tracer shutdown error", slog.Any("error", err))
		}
	}()

	// 4. Init rate limiter
	rateLimiter, err := middleware.NewMemoryRateLimiter(
		uint64(cfg.RateLimitBurst),
		time.Second, // refill window
	)
	if err != nil {
		return err
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := rateLimiter.Shutdown(shutdownCtx); err != nil {
			slog.Error("rate limiter shutdown error", slog.Any("error", err))
		}
	}()

	// 5. Build router and server
	router := server.NewRouter(cfg, rateLimiter)
	srv := server.New(cfg, router)

	// 6. Listen for OS signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine so we can block on the signal channel
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Start()
	}()

	// Block until signal or server error
	select {
	case err := <-serverErr:
		return err
	case sig := <-quit:
		slog.Info("shutdown signal received", slog.String("signal", sig.String()))
	}

	// Graceful shutdown
	if err := srv.Shutdown(); err != nil {
		return err
	}

	return nil
}