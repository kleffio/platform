package domain

// APIResponse is the standard envelope returned by all Kleff API endpoints.
type APIResponse[T any] struct {
	Data    T       `json:"data"`
	Message *string `json:"message,omitempty"`
}

// APIErrorResponse is the standard error envelope.
type APIErrorResponse struct {
	Error   string              `json:"error"`
	Code    string              `json:"code"`
	Details map[string][]string `json:"details,omitempty"`
}

// PaginatedResponse wraps a slice with pagination metadata.
type PaginatedResponse[T any] struct {
	Data       []T            `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// PaginationMeta describes the current page of results.
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}
