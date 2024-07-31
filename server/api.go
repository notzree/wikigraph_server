package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
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

func Make(handler ApiHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			slog.Error("HTTP API error", "err", err.Error(), "path", r.URL.Path)
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
