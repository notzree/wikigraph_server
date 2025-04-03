package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type ApiError struct {
	StatusCode int `json:"status_code"`
	Message    any `json:"error"`
}

func (e ApiError) Error() string {
	return fmt.Sprintf("status code: %d, error: %v", e.StatusCode, e.Message)
}

func NewApiError(status int, message any) ApiError {
	return ApiError{StatusCode: status, Message: message}
}

func InvalidJSON() ApiError {
	return NewApiError(http.StatusBadRequest, fmt.Errorf("invalid JSON request data"))
}

type ApiHandlerFunc func(w http.ResponseWriter, r *http.Request) error

// EnableCORS is a middleware that adds CORS headers to all responses
func EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow any origin or specify "http://localhost:3000"
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		// Handle preflight requests (OPTIONS)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Process the actual request
		next(w, r)
	}
}

// Modify your Make function to wrap handlers with CORS
func Make(handler ApiHandlerFunc) http.HandlerFunc {
	return EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		// Check for authorization token
		token := r.Header.Get("Authorization")
		if token == "" {
			errResp := map[string]any{
				"statusCode": http.StatusUnauthorized,
				"msg":        "unauthorized: missing authorization token",
			}
			writeJSON(w, http.StatusUnauthorized, errResp)
			log.Println("HTTP API error", "err", "missing authorization token", "path", r.URL.Path)
			return
		}

		// Process the request with the handler if token is present
		if err := handler(w, r); err != nil {
			if apiErr, ok := err.(ApiError); ok {
				writeJSON(w, apiErr.StatusCode, apiErr.Message)
			} else {
				errResp := map[string]any{
					"statusCode": http.StatusInternalServerError,
					"msg":        "internal server error",
				}
				writeJSON(w, http.StatusInternalServerError, errResp)
			}
			log.Println("HTTP API error", "err", err.Error(), "path", r.URL.Path)
		}
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
