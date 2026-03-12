package tracing

// TracingComponents holds all tracing-related components
type TracingComponents struct {
	TraceContext    TraceContext
	HTTPPropagator  HTTPPropagator
	KafkaPropagator KafkaPropagator
}

// NewTracingComponents creates a complete set of tracing components
// This is the recommended way to initialize tracing for a service
func NewTracingComponents() *TracingComponents {
	tracer := NewTraceContext()

	return &TracingComponents{
		TraceContext:    tracer,
		HTTPPropagator:  NewHTTPPropagator(tracer),
		KafkaPropagator: NewKafkaPropagator(tracer),
	}
}

// NewDefaultTraceContext creates a default TraceContext implementation
func NewDefaultTraceContext() TraceContext {
	return NewTraceContext()
}

// NewDefaultHTTPPropagator creates a default HTTPPropagator with a new TraceContext
func NewDefaultHTTPPropagator() HTTPPropagator {
	return NewHTTPPropagator(NewTraceContext())
}

// NewDefaultKafkaPropagator creates a default KafkaPropagator with a new TraceContext
func NewDefaultKafkaPropagator() KafkaPropagator {
	return NewKafkaPropagator(NewTraceContext())
}
