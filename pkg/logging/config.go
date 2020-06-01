package logging

import (
	"io"
	"os"
	"strconv"
	"time"

	"github.com/caring/go-packages/pkg/logging/internal/writer"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ReportFlag is a convenience abstraction of bool pointers
// to help avoid messy multiliners when configuring the logger
type ReportFlag *bool

var (
	trueVar  = true
	falseVar = false
)

// Config encapsulates the various settings that may be applied to a logger
type Config struct {
	// The name of the logger
	LoggerName string
	// The service name
	ServiceName string
	// All levels above this will be logged to output and to kinesis (if enabled)
	LogLevel Level
	// Dev logging out puts in a format to be consumed by the console pretty-printer
	EnableDevLogging *bool
	// The name of the kinesis stream where developer monitoring logs are piped through
	KinesisStreamMonitoring string
	// The name of the kinesis stream where business insight lgs are piped through
	KinesisStreamReporting string
	// Flag to disable kinesis
	DisableKinesis *bool
}

func newDefaultConfig() *Config {
	return &Config{
		LoggerName:              "",
		ServiceName:             "",
		LogLevel:                InfoLevel,
		EnableDevLogging:        &falseVar,
		KinesisStreamMonitoring: "",
		KinesisStreamReporting:  "",
		DisableKinesis:          &trueVar,
	}
}

// mergeAndPopulateConfig starts with a default config, and populates
// it with config from the environment. Config from the environment can
// be overridden with any config input as arguments. Only non 0 values will
// overwrite the defaults
func mergeAndPopulateConfig(c *Config) (*Config, error) {
	final := newDefaultConfig()

	if c == nil {
		c = &Config{}
	}
	if c.LoggerName != "" {
		final.LoggerName = c.LoggerName
	} else if s := os.Getenv("LOG_NAME"); s != "" {
		final.LoggerName = s
	}

	if c.ServiceName != "" {
		final.ServiceName = c.ServiceName
	} else if s := os.Getenv("SERVICE_NAME"); s != "" {
		final.ServiceName = s
	}

	if c.LogLevel != 0 {
		final.LogLevel = c.LogLevel
	} else if s := os.Getenv("LOG_LEVEL"); s != "" {
		err := final.LogLevel.Set(s)
		if err != nil {
			return nil, err
		}
	}

	if c.EnableDevLogging != nil {
		final.EnableDevLogging = c.EnableDevLogging
	} else if s := os.Getenv("LOG_ENABLE_DEV"); s != "" {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return nil, err
		}
		final.EnableDevLogging = &b
	}

	if c.KinesisStreamMonitoring != "" {
		final.KinesisStreamMonitoring = c.KinesisStreamMonitoring
	} else if s := os.Getenv("LOG_STREAM_MONITORING"); s != "" {
		final.KinesisStreamMonitoring = s
	}

	if c.KinesisStreamReporting != "" {
		final.KinesisStreamReporting = c.KinesisStreamReporting
	} else if s := os.Getenv("LOG_STREAM_REPORTING"); s != "" {
		final.KinesisStreamReporting = s
	}

	if c.DisableKinesis != nil {
		final.DisableKinesis = c.DisableKinesis
	} else if s := os.Getenv("LOG_DISABLE_KINESIS"); s != "" {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return nil, err
		}
		final.DisableKinesis = &b
	}

	return final, nil
}

// spits out a zap config that has been tuned to play nicely with
// the zap-pretty pretty printing util and easy development
func newZapDevelopmentConfig() zap.Config {
	c := zap.NewDevelopmentConfig()
	c.OutputPaths = []string{"stdout"}
	c.ErrorOutputPaths = []string{"stdout"}
	// This displays log messages in a format compatable with the zap-pretty print library
	c.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	return c
}

func buildReportingCore(streamName string, enc zapcore.EncoderConfig) (zapcore.Core, io.Closer, error) {
	w, err := writer.NewKinesisWriter(streamName)
	if err != nil {
		return nil, nil, err
	}

	buf, closer := writer.Buffer(zapcore.AddSync(w), 0, 10*time.Second)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(enc),
		buf,
		zapcore.InfoLevel,
	)

	return core, closer, nil
}

func buildMonitoringCore(streamName string, enc zapcore.EncoderConfig, lvl zapcore.Level) (zapcore.Core, io.Closer, error) {
	w, err := writer.NewKinesisWriter(streamName)
	if err != nil {
		return nil, nil, err
	}

	buf, closer := writer.Buffer(zapcore.AddSync(w), 0, 10*time.Second)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(enc),
		buf,
		lvl,
	)

	return core, closer, nil
}
