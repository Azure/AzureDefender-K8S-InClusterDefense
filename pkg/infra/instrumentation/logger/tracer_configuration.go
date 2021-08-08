package instrumenation

import (
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap/zapcore"
)

// TracerConfiguration is the configuration of the tracer.
type TracerConfiguration struct {
	TracerLevel  zapcore.Level // tracerLevel the level of the logger.
	Entry        *log.Entry    // entry is needed for Tivan's instrumentation
	DefaultTrace string
}
