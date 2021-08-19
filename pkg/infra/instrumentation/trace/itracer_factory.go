package trace

// ITracerFactory interface of tracer factory
type ITracerFactory interface {
	// CreateTracer creates ITracer interface.
	CreateTracer() (tracer ITracer)
}
