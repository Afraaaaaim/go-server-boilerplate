package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Afraaaaaim/go-server-boilerplate/internal/observability"
	"github.com/Afraaaaaim/go-server-boilerplate/pkg/apierror"
)

// Recovery catches any panic in downstream handlers, logs it with a full
// stack trace, and returns a clean 500 to the client.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				requestID := RequestIDFromContext(r.Context())

				var err error
				switch v := rec.(type) {
				case error:
					err = v
				default:
					err = fmt.Errorf("panic: %v", v)
				}

				observability.LogError(r.Context(), "panic recovered", err,
					slog.String("request_id", requestID),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
				)

				apierror.InternalError(w, requestID)
			}
		}()
		next.ServeHTTP(w, r)
	})
}