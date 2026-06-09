package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/netip"
	"time"

	"github.com/sethvargo/go-limiter"
	"github.com/sethvargo/go-limiter/memorystore"

	"github.com/Afraaaaaim/go-server-boilerplate/pkg/apierror"
)

// RateLimiterStore wraps the go-limiter store so we can shut it down cleanly.
type RateLimiterStore struct {
	store limiter.Store
}

// NewMemoryRateLimiter creates an in-memory rate limiter.
// tokens = max requests per interval, interval = window duration in seconds.
// Swap this for a Redis store when you need distributed rate limiting.
func NewMemoryRateLimiter(tokens uint64, interval time.Duration) (*RateLimiterStore, error) {
	store, err := memorystore.New(&memorystore.Config{
		Tokens:   tokens,
		Interval: interval,
	})
	if err != nil {
		return nil, err
	}
	return &RateLimiterStore{store: store}, nil
}

// Shutdown cleanly closes the underlying store.
func (rl *RateLimiterStore) Shutdown(ctx context.Context) error {
	return rl.store.Close(ctx)
}

// RateLimit returns a middleware that limits requests per client.
// Key priority:
//  1. client_id from context (set by auth middleware — per API key limiting)
//  2. X-Forwarded-For header (client behind proxy)
//  3. RemoteAddr (direct connection)
func (rl *RateLimiterStore) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := resolveKey(r)
		requestID := RequestIDFromContext(r.Context())

		_, _, _, ok, err := rl.store.Take(r.Context(), key)
		if err != nil {
			slog.ErrorContext(r.Context(), "rate limiter error",
				slog.String("request_id", requestID),
				slog.String("key", key),
				slog.Any("error", err),
			)
			// Fail open — don't block the request if the limiter itself errors
			next.ServeHTTP(w, r)
			return
		}

		if !ok {
			slog.WarnContext(r.Context(), "rate limit exceeded",
				slog.String("request_id", requestID),
				slog.String("key", key),
			)
			apierror.RateLimited(w, requestID)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// resolveKey determines the rate limit key for a request.
func resolveKey(r *http.Request) string {
	// Prefer client_id from auth middleware (per-key rate limiting)
	if clientID := ClientIDFromContext(r.Context()); clientID != "" {
		return "client:" + clientID
	}

	// Fall back to X-Forwarded-For for clients behind a proxy
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return "ip:" + xff
	}

	// Last resort: parse RemoteAddr to strip the port
	if addr, err := netip.ParseAddrPort(r.RemoteAddr); err == nil {
		return "ip:" + addr.Addr().String()
	}

	return "ip:" + r.RemoteAddr
}