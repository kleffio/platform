package http

import (
	"encoding/json"
	"net/http"

	"github.com/kleff/go-common/domain"
)

// JSON writes v as an indented JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Success wraps data in the standard APIResponse envelope and writes 200 OK.
func Success[T any](w http.ResponseWriter, data T) {
	JSON(w, http.StatusOK, domain.APIResponse[T]{Data: data})
}

// Created wraps data in the standard APIResponse envelope and writes 201 Created.
func Created[T any](w http.ResponseWriter, data T) {
	JSON(w, http.StatusCreated, domain.APIResponse[T]{Data: data})
}

// NoContent writes 204 No Content with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Paginated wraps a slice + meta in the standard PaginatedResponse envelope.
func Paginated[T any](w http.ResponseWriter, data []T, meta domain.PaginationMeta) {
	JSON(w, http.StatusOK, domain.PaginatedResponse[T]{
		Data:       data,
		Pagination: meta,
	})
}

// Error writes the standard error envelope for an *AppError. If err is not
// an *AppError, it writes a generic 500 response.
func Error(w http.ResponseWriter, err error) {
	if appErr, ok := domain.AsAppError(err); ok {
		JSON(w, appErr.Status, domain.APIErrorResponse{
			Error:   appErr.Message,
			Code:    appErr.Code,
			Details: appErr.Details,
		})
		return
	}
	JSON(w, http.StatusInternalServerError, domain.APIErrorResponse{
		Error: "an unexpected error occurred",
		Code:  "internal_error",
	})
}
