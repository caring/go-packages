package logging

import (
	"os"
	"testing"
)

func Test_newDefaultConfig(t *testing.T) {
	t.Run("Initializes a new config with the correct values", func(t *testing.T) {
		c := newDefaultConfig()

		if c.LoggerName != "" {
			t.Error("Logger name was not empty")
		}
		if c.ServiceName != "" {
			t.Error("Service name was not empty")
		}
		if c.LogLevel != "INFO" {
			t.Error("Log level was not empty")
		}
		if *c.EnableDevLogging != false {
			t.Error("Dev logging was enabled by default")
		}
		if c.KinesisStreamName != "" {
			t.Error("Kinesis stream name was not empty")
		}
		if c.KinesisPartitionKey != "" {
			t.Error("Kinesis partition key was not empty")
		}
		if *c.DisableKinesis != true {
			t.Error("Kinesis was enabled by default")
		}
	})
}

func Test_mergeAndPopulateConfig(t *testing.T) {
	t.Run("Initializes a config with default values with env and input are empty", func(t *testing.T) {
		c := &Config{}
		result, err := mergeAndPopulateConfig(c)
		if err != nil {
			t.Fatal("Error when populating config from env" + err.Error())
		}

		if result.LoggerName != "" {
			t.Error("Logger name was not empty")
		}
		if result.ServiceName != "" {
			t.Error("Service name was not empty")
		}
		if result.LogLevel != "INFO" {
			t.Error("Log level was not empty")
		}
		if *result.EnableDevLogging != false {
			t.Error("Dev logging was enabled by default")
		}
		if result.KinesisStreamName != "" {
			t.Error("Kinesis stream name was not empty")
		}
		if result.KinesisPartitionKey != "" {
			t.Error("Kinesis partition key was not empty")
		}
		if *result.DisableKinesis != true {
			t.Error("Kinesis was enabled by default")
		}
	})

	os.Setenv("SERVICE_NAME", "fooservice")
	os.Setenv("LOG_NAME", "foologger")
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_ENABLE_DEV", "TRUE")
	os.Setenv("LOG_KINESIS_NAME", "kinesisstream2")
	os.Setenv("LOG_KINESIS_KEY", "shard1")
	os.Setenv("LOG_DISABLE_KINESIS", "FALSE")

	t.Run("Initializes all config from environment correctly when given an empty config object", func(t *testing.T) {
		c := &Config{}
		result, err := mergeAndPopulateConfig(c)
		if err != nil {
			t.Fatal("Error when populating config from env" + err.Error())
		}

		if result.LoggerName != "foologger" {
			t.Error("Logger name was not foologger")
		}
		if result.ServiceName != "fooservice" {
			t.Error("Service name was not fooservice")
		}
		if result.LogLevel != "DEBUG" {
			t.Error("Log level was not DEBUG")
		}
		if *result.EnableDevLogging != true {
			t.Error("Dev logging was not true")
		}
		if result.KinesisStreamName != "kinesisstream2" {
			t.Error("Kinesis stream name was not kinesisstream2")
		}
		if result.KinesisPartitionKey != "shard1" {
			t.Error("Kinesis partition key was not shard1")
		}
		if *result.DisableKinesis != false {
			t.Error("Kinesis was not disabled")
		}
	})

	t.Run("Initializes all config from environment correctly when given a populated config object", func(t *testing.T) {
		c := &Config{
			LoggerName:          "barlogger",
			ServiceName:         "barservice",
			LogLevel:            "FATAL",
			EnableDevLogging:    &falseVar,
			KinesisStreamName:   "barstream1",
			KinesisPartitionKey: "barshard",
			DisableKinesis:      &trueVar,
		}
		result, err := mergeAndPopulateConfig(c)
		if err != nil {
			t.Fatal("Error when populating config from env" + err.Error())
		}

		if result.LoggerName != "barlogger" {
			t.Error("Logger name was not barlogger")
		}
		if result.ServiceName != "barservice" {
			t.Error("Service name was not barservice")
		}
		if result.LogLevel != "FATAL" {
			t.Error("Log level was not FATAL")
		}
		if *result.EnableDevLogging != false {
			t.Error("Dev logging was not false")
		}
		if result.KinesisStreamName != "barstream1" {
			t.Error("Kinesis stream name was not barstream1")
		}
		if result.KinesisPartitionKey != "barshard" {
			t.Error("Kinesis partition key was not barshard")
		}
		if *result.DisableKinesis != true {
			t.Error("Kinesis was not enabled")
		}
	})
}
