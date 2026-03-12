package paginator

import (
	"context"
	"math"

	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// manager implements PaginatorManager with trace integration
type manager struct {
	tracer tracing.TraceContext
}

// NewManager creates a new paginator manager with trace integration
func NewManager() PaginatorManager {
	return &manager{
		tracer: tracing.NewTraceContext(),
	}
}

// NewManagerWithTracer creates a new paginator manager with custom tracer
func NewManagerWithTracer(tracer tracing.TraceContext) PaginatorManager {
	if tracer == nil {
		tracer = tracing.NewTraceContext()
	}
	return &manager{
		tracer: tracer,
	}
}

// AdjustQuery normalizes pagination parameters with trace context
func (m *manager) AdjustQuery(ctx context.Context, query *PaginateQuery) {
	// Could add trace logging here for pagination metrics
	query.Adjust()
}

// CalculateOffset calculates database offset with trace context
func (m *manager) CalculateOffset(ctx context.Context, query PaginateQuery) int64 {
	// Could add trace logging here for performance monitoring
	return query.Offset()
}

// CreatePaginator creates paginator with trace context
func (m *manager) CreatePaginator(ctx context.Context, total, count, perPage int64, currentPage int) Paginator {
	// Could add trace logging here for pagination analytics
	return Paginator{
		Total:       total,
		Count:       count,
		PerPage:     perPage,
		CurrentPage: currentPage,
	}
}

// CreateResponse creates paginator response with trace context
func (m *manager) CreateResponse(ctx context.Context, paginator Paginator) PaginatorResponse {
	// Could add trace logging here for response metrics
	return paginator.ToResponse()
}

// Adjust normalizes the pagination parameters to valid values
// Sets defaults if values are invalid and enforces maximum limit
func (p *PaginateQuery) Adjust() {
	if p.Page < 1 {
		p.Page = DefaultPage
	}

	if p.Limit < 1 {
		p.Limit = DefaultLimit
	} else if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}
}

// AdjustWithTrace normalizes pagination parameters with trace context
func (p *PaginateQuery) AdjustWithTrace(ctx context.Context) {
	// Could add trace logging here for pagination validation
	p.Adjust()
}

// Offset calculates the database offset for the current page
func (p *PaginateQuery) Offset() int64 {
	return int64((p.Page - 1)) * p.Limit
}

// OffsetWithTrace calculates offset with trace context
func (p *PaginateQuery) OffsetWithTrace(ctx context.Context) int64 {
	// Could add trace logging here for offset calculation
	return p.Offset()
}

// TotalPages calculates the total number of pages based on total items and items per page
func (p Paginator) TotalPages() int {
	if p.Total == 0 || p.PerPage == 0 {
		return 0
	}
	return int(math.Ceil(float64(p.Total) / float64(p.PerPage)))
}

// TotalPagesWithTrace calculates total pages with trace context
func (p Paginator) TotalPagesWithTrace(ctx context.Context) int {
	// Could add trace logging here for pagination metrics
	return p.TotalPages()
}

// HasNextPage checks if there is a next page available
func (p Paginator) HasNextPage() bool {
	return p.CurrentPage < p.TotalPages()
}

// HasNextPageWithTrace checks next page availability with trace context
func (p Paginator) HasNextPageWithTrace(ctx context.Context) bool {
	// Could add trace logging here for navigation analytics
	return p.HasNextPage()
}

// HasPreviousPage checks if there is a previous page available
func (p Paginator) HasPreviousPage() bool {
	return p.CurrentPage > 1
}

// HasPreviousPageWithTrace checks previous page availability with trace context
func (p Paginator) HasPreviousPageWithTrace(ctx context.Context) bool {
	// Could add trace logging here for navigation analytics
	return p.HasPreviousPage()
}

// ToResponse converts the paginator to a response format with additional calculated fields
func (p Paginator) ToResponse() PaginatorResponse {
	return PaginatorResponse{
		Total:       p.Total,
		Count:       p.Count,
		PerPage:     p.PerPage,
		CurrentPage: p.CurrentPage,
		TotalPages:  p.TotalPages(),
		HasNext:     p.HasNextPage(),
		HasPrev:     p.HasPreviousPage(),
	}
}

// ToResponseWithTrace converts to response with trace context
func (p Paginator) ToResponseWithTrace(ctx context.Context) PaginatorResponse {
	// Could add trace logging here for response generation
	return p.ToResponse()
}

// ToPaginator converts a response back to a paginator (e.g. for deserialization)
func (p PaginatorResponse) ToPaginator() Paginator {
	return Paginator{
		Total:       p.Total,
		Count:       p.Count,
		PerPage:     p.PerPage,
		CurrentPage: p.CurrentPage,
	}
}

// ToPaginatorWithTrace converts response to paginator with trace context
func (p PaginatorResponse) ToPaginatorWithTrace(ctx context.Context) Paginator {
	// Could add trace logging here for deserialization
	return p.ToPaginator()
}

// Convenience functions for backward compatibility

// NewPaginator creates a new paginator with the given parameters
func NewPaginator(total, count, perPage int64, currentPage int) Paginator {
	return Paginator{
		Total:       total,
		Count:       count,
		PerPage:     perPage,
		CurrentPage: currentPage,
	}
}

// NewPaginatorWithTrace creates paginator with trace context
func NewPaginatorWithTrace(ctx context.Context, total, count, perPage int64, currentPage int) Paginator {
	// Could add trace logging here for paginator creation
	return NewPaginator(total, count, perPage, currentPage)
}

// NewPaginateQuery creates a new paginate query with the given parameters
func NewPaginateQuery(page int, limit int64) PaginateQuery {
	query := PaginateQuery{
		Page:  page,
		Limit: limit,
	}
	query.Adjust()
	return query
}

// NewPaginateQueryWithTrace creates paginate query with trace context
func NewPaginateQueryWithTrace(ctx context.Context, page int, limit int64) PaginateQuery {
	// Could add trace logging here for query creation
	return NewPaginateQuery(page, limit)
}
