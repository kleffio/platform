package domain

import "math"

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// PageRequest holds inbound pagination parameters.
type PageRequest struct {
	Page  int `json:"page"  form:"page"`
	Limit int `json:"limit" form:"limit"`
}

// Normalise clamps page and limit to sane values.
func (p *PageRequest) Normalise() {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.Limit < 1 {
		p.Limit = DefaultLimit
	}
	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}
}

// Offset returns the SQL/query offset for this page.
func (p *PageRequest) Offset() int {
	return (p.Page - 1) * p.Limit
}

// BuildMeta constructs a PaginationMeta given the total number of records.
func (p *PageRequest) BuildMeta(total int) PaginationMeta {
	totalPages := int(math.Ceil(float64(total) / float64(p.Limit)))
	if totalPages < 1 {
		totalPages = 1
	}
	return PaginationMeta{
		Page:       p.Page,
		Limit:      p.Limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
