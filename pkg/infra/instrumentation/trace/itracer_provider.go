package trace

// ITracerProvider provides tracer. the difference between ITracerProvider and ITracerFactory is that ITracerFactory
// creates ITracer , and the ITracerProvider doesn't create tracer,it provides exists tracer in specific context.
type ITracerProvider interface {
	// GetTracer Gets a tracer with specific context. the context is according to specific method
	//(when you create the ITraceProvider you choose the struct context)
	GetTracer(context string) (tracer ITracer)
}
