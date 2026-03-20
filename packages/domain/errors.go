package domain

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is a structured application error that carries an HTTP status code,
// a machine-readable code, and a human-readable message.
type AppError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
	// Details holds per-field validation errors, keyed by field name.
	Details map[string][]string `json:"details,omitempty"`
	// Cause is the underlying error (not serialised).
	Cause error `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// ─── Constructors ─────────────────────────────────────────────────────────────

func NewBadRequest(message string, details ...map[string][]string) *AppError {
	e := &AppError{Status: http.StatusBadRequest, Code: "bad_request", Message: message}
	if len(details) > 0 {
		e.Details = details[0]
	}
	return e
}

func NewUnauthorized(message string) *AppError {
	return &AppError{Status: http.StatusUnauthorized, Code: "unauthorized", Message: message}
}

func NewForbidden(message string) *AppError {
	return &AppError{Status: http.StatusForbidden, Code: "forbidden", Message: message}
}

func NewNotFound(resource string) *AppError {
	return &AppError{
		Status:  http.StatusNotFound,
		Code:    "not_found",
		Message: fmt.Sprintf("%s not found", resource),
	}
}

func NewConflict(message string) *AppError {
	return &AppError{Status: http.StatusConflict, Code: "conflict", Message: message}
}

func NewUnprocessable(message string, details ...map[string][]string) *AppError {
	e := &AppError{Status: http.StatusUnprocessableEntity, Code: "unprocessable_entity", Message: message}
	if len(details) > 0 {
		e.Details = details[0]
	}
	return e
}

func NewInternal(cause error) *AppError {
	return &AppError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: "an unexpected error occurred",
		Cause:   cause,
	}
}

// ─── Sentinel errors ──────────────────────────────────────────────────────────

var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

// AsAppError attempts to unwrap err as *AppError. Returns (nil, false) if err
// is not an AppError.
func AsAppError(err error) (*AppError, bool) {
	var e *AppError
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
