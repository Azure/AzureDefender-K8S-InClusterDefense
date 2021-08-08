package instrumenation

import (
	instrumenation "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/logger"
	"github.com/pkg/errors"
	"os"
	"time"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	crzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// ZaprTracerFactory implementation of ITracerFactory.
type ZaprTracerFactory struct {
	Configuration instrumenation.TracerConfiguration
}

// NewZaprTracerFactory creates new TracerFactory instance.
func NewZaprTracerFactory(configuration instrumenation.TracerConfiguration) instrumenation.ITracerFactory {
	return &ZaprTracerFactory{
		Configuration: configuration,
	}
}

// CreateTracer Creates tracer
func (factory ZaprTracerFactory) CreateTracer() (tracer instrumenation.ITracer, err error) {
	//Log level encoders
	logLevelEncoders := factory.getDefaultLogLevelEncoders()
	encoder, ok := logLevelEncoders["lower"]
	if !ok {
		err = errors.New("invalid log level encoder")
		return nil, err
	}

	switch factory.Configuration.TracerLevel.CapitalString() {
	case "DEBUG":
		eCfg := zap.NewDevelopmentEncoderConfig()
		eCfg.LevelKey = "level"
		eCfg.EncodeLevel = encoder
		tracer := crzap.New(crzap.UseDevMode(true), crzap.Encoder(zapcore.NewConsoleEncoder(eCfg)))
		ctrl.SetLogger(tracer)
		klog.SetLogger(tracer)
	case "WARNING", "ERROR":
		factory.setLoggerForProduction(encoder)
	case "INFO":
		fallthrough
	default:
		eCfg := zap.NewProductionEncoderConfig()
		eCfg.LevelKey = "level"
		eCfg.EncodeLevel = encoder
		tracer := crzap.New(crzap.UseDevMode(false), crzap.Encoder(zapcore.NewJSONEncoder(eCfg)))
		ctrl.SetLogger(tracer)
		klog.SetLogger(tracer)
	}
	return tracer, nil
}

// initialize logger for production env
func (factory ZaprTracerFactory) setLoggerForProduction(encoder zapcore.LevelEncoder) {
	sink := zapcore.AddSync(os.Stderr)
	var opts []zap.Option
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.LevelKey = "level"
	encCfg.EncodeLevel = encoder
	enc := zapcore.NewJSONEncoder(encCfg)
	lvl := zap.NewAtomicLevelAt(zap.WarnLevel)
	opts = append(opts, zap.AddStacktrace(zap.ErrorLevel),
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
		}),
		zap.AddCallerSkip(1), zap.ErrorOutput(sink))
	zlog := zap.New(zapcore.NewCore(&crzap.KubeAwareEncoder{Encoder: enc, Verbose: false}, sink, lvl))
	zlog = zlog.WithOptions(opts...)
	newTracer := zapr.NewLogger(zlog)
	ctrl.SetLogger(newTracer)
	klog.SetLogger(newTracer)
}

func (factory ZaprTracerFactory) getDefaultLogLevelEncoders() map[string]zapcore.LevelEncoder {
	return map[string]zapcore.LevelEncoder{
		"lower":        zapcore.LowercaseLevelEncoder,
		"capital":      zapcore.CapitalLevelEncoder,
		"color":        zapcore.LowercaseColorLevelEncoder,
		"capitalcolor": zapcore.CapitalColorLevelEncoder,
	}
}
