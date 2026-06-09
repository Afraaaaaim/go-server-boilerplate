package apierror

import (
	"encoding/json"
	"net/http"
)

// Error codes — add more as your API grows.
const (
	CodeUnauthorized   = "UNAUTHORIZED"
	CodeForbidden      = "FORBIDDEN"
	CodeNotFound       = "NOT_FOUND"
	CodeRateLimited    = "RATE_LIMITED"
	CodeInternalError  = "INTERNAL_ERROR"
	CodeBadRequest     = "BAD_REQUEST"
)

// APIError is the standard error response body for all API errors.
type APIError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// Write writes a JSON error response to the ResponseWriter.
func Write(w http.ResponseWriter, status int, code, message, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	body := APIError{
		Code:      code,
		Message:   message,
		RequestID: requestID,
	}

	// Intentionally ignoring encode error here — if we can't write the error
	// response there's nothing meaningful left to do.
	_ = json.NewEncoder(w).Encode(body)
}

// Convenience functions for common error types.

func Unauthorized(w http.ResponseWriter, requestID string) {
	Write(w, http.StatusUnauthorized, CodeUnauthorized, "missing or invalid API key", requestID)
}

func Forbidden(w http.ResponseWriter, requestID string) {
	Write(w, http.StatusForbidden, CodeForbidden, "access denied", requestID)
}

func RateLimited(w http.ResponseWriter, requestID string) {
	w.Header().Set("Retry-After", "1")
	Write(w, http.StatusTooManyRequests, CodeRateLimited, "rate limit exceeded", requestID)
}

func InternalError(w http.ResponseWriter, requestID string) {
	Write(w, http.StatusInternalServerError, CodeInternalError, "an internal error occurred", requestID)
}

func NotFound(w http.ResponseWriter, requestID string) {
	Write(w, http.StatusNotFound, CodeNotFound, "resource not found", requestID)
}

func BadRequest(w http.ResponseWriter, message, requestID string) {
	Write(w, http.StatusBadRequest, CodeBadRequest, message, requestID)
}