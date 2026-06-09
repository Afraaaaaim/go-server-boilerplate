package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

type healthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Healthz is the liveness probe — just confirms the process is alive.
// Should never do any I/O or dependency checks.
func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// Readyz is the readiness probe — confirms the service is ready to serve traffic.
// Add dependency checks here (DB ping, cache connectivity, etc) as you extend
// the boilerplate. Return 503 if any critical dependency is unavailable.
func Readyz(w http.ResponseWriter, r *http.Request) {
	// For now mirrors liveness — extend this as dependencies are added.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthResponse{
		Status:    "ready",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}