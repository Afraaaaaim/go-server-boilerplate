package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/Afraaaaaim/go-server-boilerplate/internal/config"
	"github.com/Afraaaaaim/go-server-boilerplate/internal/handler"
	"github.com/Afraaaaaim/go-server-boilerplate/internal/middleware"
)

// publicHandler is implemented by handler packages that have public routes.
type publicHandler interface {
	RegisterPublic(r chi.Router)
}

// protectedHandler is implemented by handler packages that have protected routes.
type protectedHandler interface {
	RegisterProtected(r chi.Router)
}

func NewRouter(cfg *config.Config, rateLimiter *middleware.RateLimiterStore) http.Handler {
	r := chi.NewRouter()

	// --- Global middleware ---
	r.Use(middleware.Recovery)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(chimiddleware.Compress(5))

	// --- Public routes mounted directly on root ---
	registerPublic(r, handler.RegisterPublic)

	// --- Protected router ---
	protected := chi.NewRouter()
	protected.Use(middleware.APIKeyAuth(cfg.APIKeys))
	protected.Use(rateLimiter.RateLimit)
	registerProtected(protected, handler.RegisterProtected)

	r.Mount("/", otelhttp.NewHandler(protected, "protected",
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
	))

	return r
}

// registerPublic mounts a public registration function and logs each route.
func registerPublic(r chi.Router, fn func(chi.Router)) {
	// Wrap in a recording router to capture what gets registered
	recorder := chi.NewRouter()
	fn(recorder)

	for _, route := range recorder.Routes() {
		for method := range route.Handlers {
			slog.Info("route registered",
				slog.String("method", method),
				slog.String("path", route.Pattern),
				slog.String("auth", "none"),
			)
		}
	}

	fn(r)
}

// registerProtected mounts a protected registration function and logs each route.
func registerProtected(r chi.Router, fn func(chi.Router)) {
	recorder := chi.NewRouter()
	fn(recorder)

	for _, route := range recorder.Routes() {
		for method := range route.Handlers {
			slog.Info("route registered",
				slog.String("method", method),
				slog.String("path", route.Pattern),
				slog.String("auth", "api-key"),
			)
		}
	}

	fn(r)
}

// routeKey is used to format a route for display purposes.
func routeKey(method, path string) string {
	return fmt.Sprintf("%s %s", method, path)
}