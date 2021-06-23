package instrumentation

import (
	"flag"
	"fmt"
	"github.com/go-logr/logr"
	"os"
	"time"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	crzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

//Log level encoders
var (
	logLevelEncoders = map[string]zapcore.LevelEncoder{
		"lower":        zapcore.LowercaseLevelEncoder,
		"capital":      zapcore.CapitalLevelEncoder,
		"color":        zapcore.LowercaseColorLevelEncoder,
		"capitalcolor": zapcore.CapitalColorLevelEncoder,
	}
)

//Logger parameters
var (
	logLevel        = flag.String("log-level", "INFO", "Minimum log level. For example, DEBUG, INFO, WARNING, ERROR. Defaulted to INFO if unspecified.")
	logLevelKey     = flag.String("log-level-key", "level", "JSON key for the log level field, defaults to `level`")
	logLevelEncoder = flag.String("log-level-encoder", "lower", "Encoder for the value of the log level field. Valid values: [`lower`, `capital`, `color`, `capitalcolor`], default: `lower`")
)

//TODO Currently using GK logger
//  (https://github.com/open-policy-agent/gatekeeper/blob/bf94eb335918f8806571e28d01fb1c26a9179b2d/main.go#L119)

//InitLogger initialize logger
func InitLogger(logger logr.Logger) {
	encoder, ok := logLevelEncoders[*logLevelEncoder]
	if !ok {
		logger.Error(fmt.Errorf("invalid log level encoder: %v", *logLevelEncoder), "Invalid log level encoder")
		os.Exit(1)
	}

	switch *logLevel {
	case "DEBUG":
		eCfg := zap.NewDevelopmentEncoderConfig()
		eCfg.LevelKey = *logLevelKey
		eCfg.EncodeLevel = encoder
		logger := crzap.New(crzap.UseDevMode(true), crzap.Encoder(zapcore.NewConsoleEncoder(eCfg)))
		ctrl.SetLogger(logger)
		klog.SetLogger(logger)
	case "WARNING", "ERROR":
		setLoggerForProduction(encoder)
	case "INFO":
		fallthrough
	default:
		eCfg := zap.NewProductionEncoderConfig()
		eCfg.LevelKey = *logLevelKey
		eCfg.EncodeLevel = encoder
		logger := crzap.New(crzap.UseDevMode(false), crzap.Encoder(zapcore.NewJSONEncoder(eCfg)))
		ctrl.SetLogger(logger)
		klog.SetLogger(logger)
	}
}

// initialize logger for production env
func setLoggerForProduction(encoder zapcore.LevelEncoder) {
	sink := zapcore.AddSync(os.Stderr)
	var opts []zap.Option
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.LevelKey = *logLevelKey
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
	newlogger := zapr.NewLogger(zlog)
	ctrl.SetLogger(newlogger)
	klog.SetLogger(newlogger)
}
