package trace

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	crzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// ZaprTracerFactory implementation of ITracerFactory.
type ZaprTracerFactory struct {
	// configuration is the configuration of the zaprtracer
	configuration *TracerConfiguration
}

// NewZaprTracerFactory creates new TracerFactory instance.
func NewZaprTracerFactory(configuration *TracerConfiguration) ITracerFactory {
	return &ZaprTracerFactory{
		configuration: configuration,
	}
}

// CreateTracer Creates tracer
func (factory *ZaprTracerFactory) CreateTracer() (tracer ITracer) {
	switch factory.configuration.TracerLevel.CapitalString() {
	// In DEBUG mode the output will be written in the following format:
	//	2021-08-10T15:26:02.610+0300	INFO	setting up cert rotation
	case "DEBUG":
		tracer = crzap.New(crzap.UseDevMode(true), crzap.Encoder(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())))
	// In INFO mode the output will be weitten in the following format:
	// {"level":"info","ts":1628598455.5155077,"logger":"controller-runtime.metrics","msg":"metrics server is starting to listen","addr":":8080"}
	case "INFO":
		fallthrough
	default:
		tracer = crzap.New(crzap.UseDevMode(false), crzap.Encoder(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())))
	}
	// SetLogger sets a concrete logging implementation for all deferred Loggers and register this tracer as the main logger
	ctrl.SetLogger(tracer)
	klog.SetLogger(tracer)
	return tracer
}
