package util

import (
	"context"

	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// MapSlice applies a converter function to each element of a slice and returns a new slice with trace integration.
// If the converter returns nil, that element is skipped.
func MapSlice[T any, R any](items []*T, converter func(*T) *R) []R {
	result := make([]R, 0, len(items))
	for _, item := range items {
		if converted := converter(item); converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}

// MapSliceWithTrace applies a converter function to each element with trace context.
func MapSliceWithTrace[T any, R any](ctx context.Context, items []*T, converter func(context.Context, *T) *R) []R {
	tracer := tracing.NewTraceContext()
	if traceID := tracer.GetTraceID(ctx); traceID != "" {
		ctx = tracer.WithTraceID(ctx, traceID)
	}

	result := make([]R, 0, len(items))
	for _, item := range items {
		if converted := converter(ctx, item); converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}

// ToInterfaceSlice converts a slice of any type to []interface{}.
// Useful for SQLBoiler WhereIn and similar operations that require []interface{}.
func ToInterfaceSlice[T any](items []T) []interface{} {
	result := make([]interface{}, len(items))
	for i, item := range items {
		result[i] = item
	}
	return result
}
