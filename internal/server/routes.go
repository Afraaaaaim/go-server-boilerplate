package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/Afraaaaaim/go-server-boilerplate/internal/config"
	"github.com/Afraaaaaim/go-server-boilerplate/internal/handler"
	"github.com/Afraaaaaim/go-server-boilerplate/internal/middleware"
)

func NewRouter(cfg *config.Config, rateLimiter *middleware.RateLimiterStore) http.Handler {
	r := chi.NewRouter()

	// --- Global middleware (applied to every request) ---
	// Order matters: recovery must be first to catch panics in other middleware.
	r.Use(middleware.Recovery)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(chimiddleware.Compress(5)) // gzip responses

	// --- Public routes (no auth, no rate limiting) ---
	r.Get("/healthz", handler.Healthz)
	r.Get("/readyz", handler.Readyz)

	// --- Protected routes ---
	r.Group(func(r chi.Router) {
		r.Use(middleware.APIKeyAuth(cfg.APIKeys))
		r.Use(rateLimiter.RateLimit)

		// Wrap with OTel HTTP instrumentation so every request gets a trace span.
		// otelhttp.NewHandler wraps the entire subrouter.
		r.Mount("/api", otelhttp.NewHandler(apiRouter(), "api",
			otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		))
	})

	return r
}

// apiRouter defines all /api/* routes.
// Add your domain routes here as the project grows.
func apiRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/example", handler.Example)
	return r
}

// ensure otelhttp timeout option compiles — remove if unused
var _ = time.Second