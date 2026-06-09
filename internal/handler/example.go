package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Afraaaaaim/go-server-boilerplate/internal/middleware"
	"github.com/Afraaaaaim/go-server-boilerplate/pkg/apierror"
)

type exampleResponse struct {
	Message  string `json:"message"`
	ClientID string `json:"client_id,omitempty"`
}

func Example(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	slog.InfoContext(ctx, "example handler called",
		slog.String("request_id", requestID),
	)

	resp := exampleResponse{
		Message: "boilerplate is working",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		apierror.InternalError(w, requestID)
		return
	}
}

// RegisterProtected registers example routes on the protected router (auth required).
func RegisterProtected(r chi.Router) {
	r.Get("/api/example", Example)
}