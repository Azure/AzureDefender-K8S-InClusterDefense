package trace

// NoOpTracerProvider is implementation that does nothing of ITracerProvider
type NoOpTracerProvider struct {
}

func NewNoOpTracerProvider() *NoOpTracerProvider {
	return &NoOpTracerProvider{}
}

// GetTracer Gets a tracer with specific context. the context is according to specific method
//(when you create the ITraceProvider you choose the struct context)
func (provider *NoOpTracerProvider) GetTracer(context string) (tracer ITracer) {
	return NewNoOpTracer()
}
