# Paginator Package

The paginator package provides pagination utilities and metadata management with distributed tracing integration for SMAP services.

## Features

- **Pagination Logic**: Calculate offsets, total pages, and navigation
- **Parameter Validation**: Automatic adjustment of invalid pagination parameters
- **Trace Integration**: Automatic trace_id propagation in pagination operations
- **Response Formatting**: Convert between internal and API response formats
- **Performance Limits**: Configurable maximum limits to prevent excessive queries
- **Backward Compatibility**: Drop-in replacement for existing paginator packages

## Usage

### Basic Usage (Backward Compatible)

```go
import "github.com/smap-hcmut/shared-libs/go/paginator"

// Create pagination query
query := paginator.NewPaginateQuery(2, 20) // page 2, 20 items per page

// Adjust invalid parameters
query.Adjust() // Ensures valid page/limit values

// Calculate database offset
offset := query.Offset() // Returns 20 for page 2

// Create paginator with results
pag := paginator.NewPaginator(150, 20, 20, 2) // total=150, count=20, perPage=20, page=2

// Get pagination metadata
totalPages := pag.TotalPages()    // Returns 8
hasNext := pag.HasNextPage()      // Returns true
hasPrev := pag.HasPreviousPage()  // Returns true

// Convert to API response format
response := pag.ToResponse()
```

### Advanced Usage with Trace Integration

```go
import (
    "github.com/smap-hcmut/shared-libs/go/paginator"
    "context"
)

// Create manager with trace integration
manager := paginator.NewManager()

// All operations with trace context
query := paginator.NewPaginateQueryWithTrace(ctx, 2, 20)
manager.AdjustQuery(ctx, &query)
offset := manager.CalculateOffset(ctx, query)

pag := manager.CreatePaginator(ctx, 150, 20, 20, 2)
response := manager.CreateResponse(ctx, pag)

// Trace-aware methods
totalPages := pag.TotalPagesWithTrace(ctx)
hasNext := pag.HasNextPageWithTrace(ctx)
response := pag.ToResponseWithTrace(ctx)
```

### HTTP Handler Integration

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/smap-hcmut/shared-libs/go/paginator"
)

type ListRequest struct {
    paginator.PaginateQuery
    Search string `form:"search"`
}

func ListUsers(c *gin.Context) {
    var req ListRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // Adjust pagination parameters
    req.PaginateQuery.Adjust()
    
    // Calculate offset for database query
    offset := req.PaginateQuery.Offset()
    
    // Query database
    users, total, err := userService.List(ctx, req.Search, offset, req.Limit)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    // Create pagination metadata
    pag := paginator.NewPaginator(total, int64(len(users)), req.Limit, req.Page)
    
    // Return paginated response
    c.JSON(200, gin.H{
        "users":      users,
        "pagination": pag.ToResponse(),
    })
}
```

### Database Integration

```go
// Repository method with pagination
func (r *UserRepository) List(ctx context.Context, search string, offset, limit int64) ([]User, int64, error) {
    // Count total records
    var total int64
    countQuery := r.db.Model(&User{})
    if search != "" {
        countQuery = countQuery.Where("username ILIKE ?", "%"+search+"%")
    }
    if err := countQuery.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // Get paginated records
    var users []User
    query := r.db.Model(&User{}).Offset(int(offset)).Limit(int(limit))
    if search != "" {
        query = query.Where("username ILIKE ?", "%"+search+"%")
    }
    if err := query.Find(&users).Error; err != nil {
        return nil, 0, err
    }
    
    return users, total, nil
}

// Service method with pagination
func (s *UserService) List(ctx context.Context, search string, offset, limit int64) ([]User, int64, error) {
    // Use trace-aware pagination if needed
    manager := paginator.NewManager()
    
    // Could add pagination metrics here
    users, total, err := s.repo.List(ctx, search, offset, limit)
    if err != nil {
        return nil, 0, err
    }
    
    return users, total, nil
}
```

### Service-to-Service Communication

```go
// Client-side: Send pagination parameters
func (c *APIClient) ListUsers(ctx context.Context, page int, limit int64) (*UserListResponse, error) {
    query := paginator.NewPaginateQuery(page, limit)
    
    req, err := http.NewRequestWithContext(ctx, "GET", "/users", nil)
    if err != nil {
        return nil, err
    }
    
    // Add pagination parameters
    q := req.URL.Query()
    q.Set("page", fmt.Sprintf("%d", query.Page))
    q.Set("limit", fmt.Sprintf("%d", query.Limit))
    req.URL.RawQuery = q.Encode()
    
    // Make request...
}

// Server-side: Parse and validate pagination
func HandleListUsers(c *gin.Context) {
    var query paginator.PaginateQuery
    if err := c.ShouldBindQuery(&query); err != nil {
        c.JSON(400, gin.H{"error": "Invalid pagination parameters"})
        return
    }
    
    // Adjust ensures valid values
    query.Adjust()
    
    // Use in service call...
}
```

## API Reference

### Types

#### PaginateQuery
Request pagination parameters:
```go
type PaginateQuery struct {
    Page  int   `json:"page" form:"page"`   // Page number (1-indexed)
    Limit int64 `json:"limit" form:"limit"` // Items per page
}
```

#### Paginator
Internal pagination metadata:
```go
type Paginator struct {
    Total       int64 `json:"total"`        // Total items
    Count       int64 `json:"count"`        // Items in current page
    PerPage     int64 `json:"per_page"`     // Items per page
    CurrentPage int   `json:"current_page"` // Current page (1-indexed)
}
```

#### PaginatorResponse
API response format:
```go
type PaginatorResponse struct {
    Total       int64 `json:"total"`        // Total items
    Count       int64 `json:"count"`        // Items in current page
    PerPage     int64 `json:"per_page"`     // Items per page
    CurrentPage int   `json:"current_page"` // Current page
    TotalPages  int   `json:"total_pages"`  // Total pages
    HasNext     bool  `json:"has_next"`     // Has next page
    HasPrev     bool  `json:"has_prev"`     // Has previous page
}
```

### Methods

#### PaginateQuery Methods
- `Adjust()`: Normalize parameters to valid values
- `AdjustWithTrace(ctx)`: Adjust with trace context
- `Offset() int64`: Calculate database offset
- `OffsetWithTrace(ctx) int64`: Calculate offset with trace context

#### Paginator Methods
- `TotalPages() int`: Calculate total number of pages
- `TotalPagesWithTrace(ctx) int`: Calculate with trace context
- `HasNextPage() bool`: Check if next page exists
- `HasNextPageWithTrace(ctx) bool`: Check with trace context
- `HasPreviousPage() bool`: Check if previous page exists
- `HasPreviousPageWithTrace(ctx) bool`: Check with trace context
- `ToResponse() PaginatorResponse`: Convert to API response format
- `ToResponseWithTrace(ctx) PaginatorResponse`: Convert with trace context

#### Constructor Functions
- `NewPaginateQuery(page, limit) PaginateQuery`: Create query
- `NewPaginateQueryWithTrace(ctx, page, limit) PaginateQuery`: Create with trace
- `NewPaginator(total, count, perPage, currentPage) Paginator`: Create paginator
- `NewPaginatorWithTrace(ctx, ...) Paginator`: Create with trace context

#### Manager Interface
- `NewManager() PaginatorManager`: Create manager with default tracer
- `NewManagerWithTracer(tracer) PaginatorManager`: Create with custom tracer

### Manager Methods
- `AdjustQuery(ctx, query)`: Adjust query parameters with trace
- `CalculateOffset(ctx, query) int64`: Calculate offset with trace
- `CreatePaginator(ctx, ...) Paginator`: Create paginator with trace
- `CreateResponse(ctx, paginator) PaginatorResponse`: Create response with trace

## Constants

- `DefaultPage`: Default page number (1)
- `DefaultLimit`: Default items per page (15)
- `MaxLimit`: Maximum items per page (100)

## Validation Rules

### Page Validation
- Invalid page (< 1) → Set to `DefaultPage` (1)
- Valid range: 1 to unlimited

### Limit Validation
- Invalid limit (< 1) → Set to `DefaultLimit` (15)
- Excessive limit (> MaxLimit) → Set to `MaxLimit` (100)
- Valid range: 1 to 100

## Migration Guide

### From Local Paginator Package

1. Update imports:
```go
// Before
import "your-service/pkg/paginator"

// After
import "github.com/smap-hcmut/shared-libs/go/paginator"
```

2. No code changes needed for basic usage
3. Optional: Add trace integration for enhanced monitoring

### Response Format Changes

The package maintains backward compatibility:
```go
// Old format still works
response := paginator.ToResponse()

// New trace-aware format available
response := paginator.ToResponseWithTrace(ctx)
```

### Trace Integration Benefits

- **Performance Monitoring**: Track pagination query performance
- **Usage Analytics**: Monitor pagination patterns across services
- **Debugging**: Easier troubleshooting with trace context
- **Optimization**: Identify inefficient pagination usage

## Best Practices

1. **Always Adjust**: Call `Adjust()` on user input to ensure valid parameters
2. **Limit Enforcement**: Use `MaxLimit` to prevent excessive database queries
3. **Offset Calculation**: Use `Offset()` method for consistent database queries
4. **Response Format**: Use `ToResponse()` for consistent API responses
5. **Trace Integration**: Use trace-aware methods for better observability
6. **Error Handling**: Validate pagination parameters in HTTP handlers
7. **Performance**: Consider caching total counts for large datasets

## Performance Considerations

- **Database Queries**: Large offsets can be slow; consider cursor-based pagination for large datasets
- **Total Counts**: Counting large tables can be expensive; consider approximations
- **Memory Usage**: Large page sizes increase memory usage
- **Network**: Large responses increase network overhead

## Example Response

```json
{
  "users": [...],
  "pagination": {
    "total": 150,
    "count": 20,
    "per_page": 20,
    "current_page": 2,
    "total_pages": 8,
    "has_next": true,
    "has_prev": true
  }
}
```