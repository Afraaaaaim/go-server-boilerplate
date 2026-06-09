package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// responseWriter wraps http.ResponseWriter to capture the status code
// written by downstream handlers.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

// RequestID injects a unique request ID into the context and response headers.
// Always place this first in the middleware chain.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respect an existing ID forwarded by an upstream proxy/gateway
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), contextKeyRequestID, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logger logs each incoming request with method, path, status, latency,
// request ID, and client ID. Uses slog so OTel trace context propagates.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := wrapResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		latency := time.Since(start)
		status := wrapped.status
		requestID := RequestIDFromContext(r.Context())
		clientID := ClientIDFromContext(r.Context())

		attrs := []any{
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_id", clientID),
			slog.String("user_agent", r.UserAgent()),
			slog.String("remote_addr", r.RemoteAddr),
		}

		switch {
		case status >= 500:
			slog.ErrorContext(r.Context(), "request completed", attrs...)
		case status >= 400:
			slog.WarnContext(r.Context(), "request completed", attrs...)
		default:
			slog.InfoContext(r.Context(), "request completed", attrs...)
		}
	})
}