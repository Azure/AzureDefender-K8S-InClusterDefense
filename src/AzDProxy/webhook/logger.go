package webhook

import (
	"fmt"
	"os"
	"time"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	crzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	SetupLog         = ctrl.Log.WithName("setup")
	logLevelEncoders = map[string]zapcore.LevelEncoder{
		"lower":        zapcore.LowercaseLevelEncoder,
		"capital":      zapcore.CapitalLevelEncoder,
		"color":        zapcore.LowercaseColorLevelEncoder,
		"capitalcolor": zapcore.CapitalColorLevelEncoder,
	}
)

func InitLogger() {
	encoder, ok := logLevelEncoders[*logLevelEncoder]
	if !ok {
		SetupLog.Error(fmt.Errorf("invalid log level encoder: %v", *logLevelEncoder), "Invalid log level encoder")
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
