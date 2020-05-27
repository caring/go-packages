package logging

import (
	"os"
	"strconv"
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
	// The name of the kinesis stream
	KinesisStreamName string
	// The partition key to determine which kinesis shard to write to
	KinesisPartitionKey string
	// Flag to disable kinesis
	DisableKinesis *bool
}

var (
	trueVar  = true
	falseVar = false
)

func newDefaultConfig() *Config {
	return &Config{
		LoggerName:          "",
		ServiceName:         "",
		LogLevel:            InfoLevel,
		EnableDevLogging:    &falseVar,
		KinesisStreamName:   "",
		KinesisPartitionKey: "",
		DisableKinesis:      &trueVar,
	}
}

// mergeAndPopulateConfig starts with a default config, and populates
// it with config from the environment. Config from the environment can
// be overridden with any config input as arguments. Only non 0 values will
// overwrite the defaults
func mergeAndPopulateConfig(c *Config) (*Config, error) {
	final := newDefaultConfig()

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

	if c.KinesisStreamName != "" {
		final.KinesisStreamName = c.KinesisStreamName
	} else if s := os.Getenv("LOG_KINESIS_NAME"); s != "" {
		final.KinesisStreamName = s
	}

	if c.KinesisPartitionKey != "" {
		final.KinesisPartitionKey = c.KinesisPartitionKey
	} else if s := os.Getenv("LOG_KINESIS_KEY"); s != "" {
		final.KinesisPartitionKey = s
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
