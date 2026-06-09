package handler

import (
    "encoding/json"
    "log/slog"
    "net/http"

    "github.com/Afraaaaaim/go-server-boilerplate/internal/middleware"
    "github.com/Afraaaaaim/go-server-boilerplate/pkg/apierror"
)

type exampleResponse struct {
	Message  string `json:"message"`
	ClientID string `json:"client_id,omitempty"`
}

// Example is a placeholder handler showing the standard pattern:
// read context values, do work, write JSON response.
// Replace or delete this once you add real domain handlers.
func Example(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Pull request-scoped values injected by middleware
	requestID := middleware.RequestIDFromContext(ctx)  // from context, not header

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