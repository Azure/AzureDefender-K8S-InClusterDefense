package trace

import (
	"go.uber.org/zap/zapcore"
)

// Encoder is enum that indicate how the tracer should encode the msgs.
type Encoder int

const (
	// NONE no encoding.
	NONE Encoder = iota
	// JSON encode msg to json.
	JSON
)

// TracerConfiguration is the configuration of the tracer.
type TracerConfiguration struct {
	// TracerLevel is the level of the logger.
	TracerLevel zapcore.Level
	// DefaultTrace
	DefaultContext string
	// EncoderLogs
	EncoderLogs Encoder
}
