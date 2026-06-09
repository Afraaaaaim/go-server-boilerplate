package observability

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/log/global"

	"github.com/Afraaaaaim/go-server-boilerplate/internal/config"
)

// InitLogger sets up the global slog logger based on config.
// Call this once at startup before anything else logs.
func InitLogger(cfg *config.Config) error {
	level, err := parseLevel(cfg.LogLevel)
	if err != nil {
		return err
	}

	// Determine output destination
	var output io.Writer = os.Stdout
	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file %q: %w", cfg.LogFile, err)
		}
		output = f
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true, // adds filename, function, line number to every record
	}

	// Build base handler based on format flag
	var baseHandler slog.Handler
	switch cfg.LogFormat {
	case "json":
		baseHandler = slog.NewJSONHandler(output, opts)
	default:
		baseHandler = slog.NewTextHandler(output, opts)
	}

	// If OTLP is enabled, wrap with the otelslog bridge so every log record
	// also flows through the OTel Logs SDK with trace/span IDs attached.
	var handler slog.Handler = baseHandler
	if cfg.OTLPEnabled() {
		otelHandler := otelslog.NewHandler(
			cfg.ServiceName,
			otelslog.WithLoggerProvider(global.GetLoggerProvider()),
		)
		handler = &multiHandler{handlers: []slog.Handler{baseHandler, otelHandler}}
	}

	logger := slog.New(handler).With(
		slog.String("service", cfg.ServiceName),
		slog.String("env", cfg.Env),
	)

	slog.SetDefault(logger)
	return nil
}

// LogError logs an error with a full stack trace, the error value,
// and any extra attributes you pass in.
func LogError(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	args := []any{
		slog.Any("error", err),
		slog.String("stack", captureStack(2)),
	}
	for _, a := range attrs {
		args = append(args, a)
	}
	slog.ErrorContext(ctx, msg, args...)
}

// parseLevel converts a string level to slog.Level.
func parseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level %q", s)
	}
}

// captureStack returns a formatted stack trace string, skipping `skip` frames.
func captureStack(skip int) string {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip+2, pcs)
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	var sb strings.Builder
	for {
		frame, more := frames.Next()
		// Skip runtime internals
		if strings.Contains(frame.File, "runtime/") {
			if !more {
				break
			}
			continue
		}
		fmt.Fprintf(&sb, "\n  %s\n    %s:%d", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return sb.String()
}

// multiHandler fans out a single log record to multiple slog.Handler instances.
// Used to write to both a local handler (text/json) and the OTel bridge simultaneously.
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var lastErr error
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				lastErr = err
			}
		}
	}
	return lastErr
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}