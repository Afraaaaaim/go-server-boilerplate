package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/Afraaaaaim/go-server-boilerplate/pkg/apierror"
)

// contextKey is an unexported type for context keys in this package.
// Prevents collisions with keys from other packages.
type contextKey string

const (
	contextKeyClientID  contextKey = "client_id"
	contextKeyRequestID contextKey = "request_id"
)

// ClientIDFromContext retrieves the client ID injected by the auth middleware.
func ClientIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(contextKeyClientID).(string)
	return v
}

// RequestIDFromContext retrieves the request ID injected by the RequestID middleware.
func RequestIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(contextKeyRequestID).(string)
	return v
}

// APIKeyAuth returns a middleware that validates API keys.
// It accepts both:
//
//	Authorization: Bearer <key>
//	X-API-Key: <key>
//
// If validKeys is empty, the middleware is skipped (auth disabled).
func APIKeyAuth(validKeys []string) func(http.Handler) http.Handler {
	// Pre-hash all valid keys at startup so we never compare raw keys at runtime.
	hashed := make(map[string]struct{}, len(validKeys))
	for _, k := range validKeys {
		hashed[hashKey(k)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Auth disabled — pass through
			if len(hashed) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			key := extractKey(r)
			if key == "" {
				requestID := RequestIDFromContext(r.Context())
				apierror.Unauthorized(w, requestID)
				return
			}

			h := hashKey(key)
			if _, ok := hashed[h]; !ok {
				requestID := RequestIDFromContext(r.Context())
				apierror.Unauthorized(w, requestID)
				return
			}

			// Inject a client ID (first 8 chars of the hash) into context
			// so downstream middleware and handlers can identify the caller
			// without ever touching the raw key.
			clientID := h[:8]
			ctx := context.WithValue(r.Context(), contextKeyClientID, clientID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractKey pulls the API key from the request headers.
func extractKey(r *http.Request) string {
	// Try Authorization: Bearer <key> first
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}
	// Fall back to X-API-Key header
	return r.Header.Get("X-API-Key")
}

// hashKey returns the SHA-256 hex digest of a key.
func hashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}