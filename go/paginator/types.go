package paginator

import "context"

// PaginateQuery contains pagination parameters for a request with trace integration
type PaginateQuery struct {
	Page  int   `json:"page" form:"page"`   // Page number (1-indexed)
	Limit int64 `json:"limit" form:"limit"` // Number of items per page
}

// Paginator contains pagination metadata for a query result with trace integration
type Paginator struct {
	Total       int64 `json:"total"`        // Total number of items across all pages
	Count       int64 `json:"count"`        // Number of items in current page
	PerPage     int64 `json:"per_page"`     // Number of items per page
	CurrentPage int   `json:"current_page"` // Current page number (1-indexed)
}

// PaginatorResponse is the response format for pagination metadata
type PaginatorResponse struct {
	Total       int64 `json:"total"`        // Total number of items across all pages
	Count       int64 `json:"count"`        // Number of items in current page
	PerPage     int64 `json:"per_page"`     // Number of items per page
	CurrentPage int   `json:"current_page"` // Current page number (1-indexed)
	TotalPages  int   `json:"total_pages"`  // Total number of pages
	HasNext     bool  `json:"has_next"`     // Whether there is a next page
	HasPrev     bool  `json:"has_prev"`     // Whether there is a previous page
}

// PaginatorManager provides pagination utilities with trace integration
type PaginatorManager interface {
	// AdjustQuery normalizes pagination parameters with trace context
	AdjustQuery(ctx context.Context, query *PaginateQuery)
	// CalculateOffset calculates database offset with trace context
	CalculateOffset(ctx context.Context, query PaginateQuery) int64
	// CreatePaginator creates paginator with trace context
	CreatePaginator(ctx context.Context, total, count, perPage int64, currentPage int) Paginator
	// CreateResponse creates paginator response with trace context
	CreateResponse(ctx context.Context, paginator Paginator) PaginatorResponse
}
