package trace

// TracerProvider implements ITracerProvider interface. it wraps exists ITracer and add to
//ITracer context in the struct level. e.g. Server contains TracerProvider struct with the "Server" context.
type TracerProvider struct {
	// tracer of the TracerProvider.
	tracer ITracer
}

// NewTracerProvider  gets an exists ITracer and context, and it wraps the tracer with the new context
func NewTracerProvider(tracer ITracer, context string) (provider ITracerProvider) {
	return &TracerProvider{
		tracer: tracer.WithName(context),
	}
}

// GetTracer returns tracer with new context (the context now is in the method level - e.g. Server.Run is the context)
func (provider *TracerProvider) GetTracer(context string) (tracer ITracer) {
	return provider.tracer.WithName(context)
}
