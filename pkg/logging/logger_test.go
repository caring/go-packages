package logging

import (
	"testing"

	"github.com/caring/go-packages/pkg/logging/exit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

var config = &Config{
	LogLevel: DebugLevel,
}

func Test_LoggerNewChild(t *testing.T) {
	withLogger(config, func(logger *Logger, logs *observer.ObservedLogs) {
		logger.With(nil, Int64("foo", 42))
		// Child loggers should have copy-on-write semantics, so two children
		// shouldn't stomp on each other's fields or affect the parent's fields.
		logger.NewChild(nil, String("one", "two")).Info("")
		logger.NewChild(nil, String("three", "four")).Info("")
		logger.Info("")

		assert.Equal(t, []observer.LoggedEntry{
			{Context: commonFields(config, FieldOpts{}, zap.Int64("foo", 42), zap.String("one", "two"))},
			{Context: commonFields(config, FieldOpts{}, zap.Int64("foo", 42), zap.String("three", "four"))},
			{Context: commonFields(config, FieldOpts{}, zap.Int64("foo", 42))},
		}, logs.AllUntimed(), "Unexpected cross-talk between child loggers.")
	})
}

func Test_LoggerSetInternalFields(t *testing.T) {
	t.Run("Overwriting existing values with non 0 values", func(t *testing.T) {
		l := &Logger{
			endpoint: "someendpoint",
			clientID: "someclient",
		}

		l.With(&FieldOpts{
			Endpoint: "newendpoint",
			ClientID: "NewClient",
			UserID:   "someuser",
		})

		assert.Equal(t, "newendpoint", l.endpoint, "Expected endpoint to change")
		assert.Equal(t, "NewClient", l.clientID, "Expected client ID to change")
		assert.Equal(t, "someuser", l.userID, "Expected user ID to change")
		assert.Equal(t, "", l.correlationID, "Expected correlation ID not to change")
	})

	t.Run("Resets fields when options order it", func(t *testing.T) {
		l := &Logger{
			endpoint:      "someendpoint",
			clientID:      "someclient",
			correlationID: "someID",
		}

		l.With(&FieldOpts{
			ResetEndpoint: true,
			ResetClientID: true,
		})

		assert.Equal(t, "", l.endpoint, "Expected endpoint be reset")
		assert.Equal(t, "", l.clientID, "Expected client ID be reset")
		assert.Equal(t, "someID", l.correlationID, "Expected correlation ID not to change")
	})

	t.Run("Accumulates fields on existing ones", func(t *testing.T) {
		s := String("one", "two")
		l := &Logger{
			fields: []Field{s},
		}

		ss := String("three", "four")
		l.With(nil, ss)

		require.Equal(t, 2, len(l.fields), "Expected fields to be accumulated on the logger")
		assert.Equal(t, s, l.fields[0], "Expected existing fields not to be overwritten")
		assert.Equal(t, ss, l.fields[1], "Expected the second field to be accumulated")
	})

	t.Run("Overwrites accumulated fields when the options require", func(t *testing.T) {
		l := &Logger{
			fields: []Field{String("one", "two")},
		}

		ss := String("three", "four")
		l.With(&FieldOpts{OverwriteAccumulatedFields: true}, ss)

		require.Equal(t, 1, len(l.fields), "Expected fields to be accumulated on the logger")
		assert.Equal(t, ss, l.fields[0], "Expected the second field to be accumulated")
	})
}

func Test_LoggerLogPanic(t *testing.T) {

	withLogger(config, func(logger *Logger, logs *observer.ObservedLogs) {
		assert.Panics(t, func() { logger.Panic("baz") }, "Expected panic")

		output := logs.AllUntimed()
		assert.Equal(t, 7, len(output[0].Context), "Unexpected context on first log.")
		assert.Equal(
			t,
			zapcore.Entry{Message: "baz", Level: zap.PanicLevel},
			output[0].Entry,
			"Unexpected output from panic-level Log.",
		)
	})
}

func Test_LoggerLogFatal(t *testing.T) {
	withLogger(config, func(logger *Logger, logs *observer.ObservedLogs) {
		stub := exit.WithStub(func() {
			logger.Fatal("baz")
		})
		assert.True(t, stub.Exited, "Expected Fatal logger call to terminate process.")
		output := logs.AllUntimed()
		assert.Equal(t, 7, len(output[0].Context), "Unexpected context on first log.")
		assert.Equal(
			t,
			zapcore.Entry{Message: "baz", Level: zap.FatalLevel},
			output[0].Entry,
			"Unexpected output from fatal-level Log.",
		)
	})

}

func Test_LoggerLeveledMethods(t *testing.T) {
	withLogger(config, func(logger *Logger, logs *observer.ObservedLogs) {
		tests := []struct {
			method        func(string, ...Field)
			expectedLevel zapcore.Level
		}{
			{logger.Debug, zap.DebugLevel},
			{logger.Info, zap.InfoLevel},
			{logger.Warn, zap.WarnLevel},
			{logger.Error, zap.ErrorLevel},
		}
		for i, tt := range tests {
			tt.method("")
			output := logs.AllUntimed()
			assert.Equal(t, i+1, len(output), "Unexpected number of logs.")
			assert.Equal(t, 7, len(output[i].Context), "Unexpected context on first log.")
			assert.Equal(
				t,
				zapcore.Entry{Level: tt.expectedLevel},
				output[i].Entry,
				"Unexpected output from %s-level logger method.", tt.expectedLevel)
		}
	})
}

func Test_LoggerAlwaysPanics(t *testing.T) {
	// Users can disable writing out panic-level logs, but calls to logger.Panic()
	// should still call panic().
	withLogger(&Config{LogLevel: FatalLevel}, func(logger *Logger, logs *observer.ObservedLogs) {
		msg := "Even if output is disabled, logger.Panic should always panic."
		assert.Panics(t, func() { logger.Panic("foo") }, msg)
		assert.Equal(t, 0, logs.Len(), "Panics shouldn't be written out if PanicLevel is disabled.")
	})
}

func Test_LoggerAlwaysFatals(t *testing.T) {
	// Users can disable writing out fatal-level logs, but calls to logger.Fatal()
	// should still terminate the process.
	withLogger(&Config{LogLevel: FatalLevel + 1}, func(logger *Logger, logs *observer.ObservedLogs) {
		stub := exit.WithStub(func() { logger.Fatal("") })
		assert.True(t, stub.Exited, "Expected calls to logger.Fatal to terminate process.")
		assert.Equal(t, 0, logs.Len(), "Shouldn't write out logs when fatal-level logging is disabled.")
	})
}

func Test_LoggerDPanic(t *testing.T) {
	withLogger(config, func(logger *Logger, logs *observer.ObservedLogs) {
		assert.NotPanics(t, func() { logger.DPanic("") })
		assert.Equal(
			t,
			[]observer.LoggedEntry{
				{
					Entry: zapcore.Entry{Level: zap.DPanicLevel},
					Context: commonFields(
						config,
						FieldOpts{},
					),
				},
			},
			logs.AllUntimed(),
			"Unexpected log output from DPanic in production mode.",
		)
	})
	c := &Config{
		EnableDevLogging: &trueVar,
		LogLevel:         DPanicLevel,
	}
	withLogger(c, func(logger *Logger, logs *observer.ObservedLogs) {
		assert.Panics(t, func() { logger.DPanic("") })
		assert.Equal(
			t,
			[]observer.LoggedEntry{
				{
					Entry: zapcore.Entry{Level: zap.DPanicLevel},
					Context: commonFields(
						c,
						FieldOpts{},
					),
				},
			},
			logs.AllUntimed(),
			"Unexpected log output from DPanic in development mode.",
		)
	})
}

func Test_LoggerNoOpsDisabledLevels(t *testing.T) {
	withLogger(&Config{LogLevel: WarnLevel}, func(logger *Logger, logs *observer.ObservedLogs) {
		logger.Info("silence!")
		assert.Equal(
			t,
			[]observer.LoggedEntry{},
			logs.AllUntimed(),
			"Expected logging at a disabled level to produce no output.",
		)
	})
}
