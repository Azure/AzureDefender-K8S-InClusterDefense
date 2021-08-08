package instrumenation

// ITracerFactory interface of tracer factory
type ITracerFactory interface {
	CreateTracer() (tracer ITracer, err error)
}
