package logging

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newDefaultConfig(t *testing.T) {
	c := newDefaultConfig()

	assert.Equal(t, "", c.LoggerName, "Expected an empty logger name")
	assert.Equal(t, "", c.ServiceName, "Expected an empty service name")
	assert.Equal(t, InfoLevel, c.LogLevel, "Expected INFO log level")
	assert.Equal(t, false, *c.EnableDevLogging, "Expected dev logging to be disabled")
	assert.Equal(t, "", c.KinesisStreamMonitoring, "Expected blank kinesis stream")
	assert.Equal(t, "", c.KinesisStreamReporting, "Expected blank kinesis stream")
	assert.Equal(t, true, *c.DisableKinesis, "Expected kinesis to be disabled")
}

func Test_mergeAndPopulateConfig(t *testing.T) {
	t.Run("Initializes a config with default values with env and input are empty", func(t *testing.T) {
		c := &Config{}
		result, err := mergeAndPopulateConfig(c)

		require.NoError(t, err, "Expected no error creating config")
		assert.Equal(t, "", result.LoggerName, "Expected an empty logger name")
		assert.Equal(t, "", result.ServiceName, "Expected an empty service name")
		assert.Equal(t, InfoLevel, result.LogLevel, "Expected INFO log level")
		assert.Equal(t, false, *result.EnableDevLogging, "Expected dev logging to be disabled")
		assert.Equal(t, "", result.KinesisStreamMonitoring, "Expected blank kinesis stream")
		assert.Equal(t, "", result.KinesisStreamReporting, "Expected blank kinesis stream")
		assert.Equal(t, true, *result.DisableKinesis, "Expected kinesis to be disabled")
	})

	os.Setenv("SERVICE_NAME", "fooservice")
	os.Setenv("LOG_NAME", "foologger")
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_ENABLE_DEV", "TRUE")
	os.Setenv("LOG_STREAM_MONITORING", "monitoringstream2")
	os.Setenv("LOG_STREAM_REPORTING", "reportingstream2")
	os.Setenv("LOG_DISABLE_KINESIS", "FALSE")

	t.Run("Initializes all config from environment correctly when given an empty config object", func(t *testing.T) {
		c := &Config{}
		result, err := mergeAndPopulateConfig(c)

		require.NoError(t, err, "Expected no error creating config")
		assert.Equal(t, "foologger", result.LoggerName, "Expected logger name to be foologger")
		assert.Equal(t, "fooservice", result.ServiceName, "Expected service name to be fooservice")
		assert.Equal(t, DebugLevel, result.LogLevel, "Expected DEBUG log level")
		assert.Equal(t, true, *result.EnableDevLogging, "Expected dev logging to be enabled")
		assert.Equal(t, "monitoringstream2", result.KinesisStreamMonitoring, "Expected stream name to  kinesisstream2")
		assert.Equal(t, "reportingstream2", result.KinesisStreamReporting, "Expected blank kinesis shard to be shard1")
		assert.Equal(t, false, *result.DisableKinesis, "Expected kinesis to be enabled")
	})

	t.Run("Initializes all config from environment correctly when given a populated config object", func(t *testing.T) {
		c := &Config{
			LoggerName:              "barlogger",
			ServiceName:             "barservice",
			LogLevel:                FatalLevel,
			EnableDevLogging:        &falseVar,
			KinesisStreamMonitoring: "barmonitor1",
			KinesisStreamReporting:  "barreport1",
			DisableKinesis:          &trueVar,
		}
		result, err := mergeAndPopulateConfig(c)

		require.NoError(t, err, "Expected no error creating config")
		assert.Equal(t, "barlogger", result.LoggerName, "Expected logger name to be barlogger")
		assert.Equal(t, "barservice", result.ServiceName, "Expected service name to be barservice")
		assert.Equal(t, FatalLevel, result.LogLevel, "Expected FATAL log level")
		assert.Equal(t, false, *result.EnableDevLogging, "Expected dev logging to be disabled")
		assert.Equal(t, "barmonitor1", result.KinesisStreamMonitoring, "Expected stream name to  barstream1")
		assert.Equal(t, "barreport1", result.KinesisStreamReporting, "Expected blank kinesis shard to be barshard")
		assert.Equal(t, true, *result.DisableKinesis, "Expected kinesis to be disabled")
	})

	os.Setenv("SERVICE_NAME", "")
	os.Setenv("LOG_NAME", "")
	os.Setenv("LOG_LEVEL", "")
	os.Setenv("LOG_ENABLE_DEV", "")
	os.Setenv("LOG_STREAM_MONITORING", "")
	os.Setenv("LOG_STREAM_REPORTING", "")
	os.Setenv("LOG_DISABLE_KINESIS", "")
}
