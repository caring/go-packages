package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// A utility method ported from Uber's zap lib to aid in testing log outputs
func withLogger(c *Config, f func(*Logger, *observer.ObservedLogs)) {
	l := zapcore.Level(c.LogLevel)
	fac, logs := observer.New(l)

	opts := []zap.Option{}

	if b := c.EnableDevLogging; b != nil && *b {
		opts = append(opts, zap.Development())
	}
	zapL := zap.New(fac, opts...)

	log, err := NewLogger(c)
	if err != nil {
		panic(err)
	}
	log.internalLogger = zapL
	f(log, logs)
}

// A testing util that creates an array of zap fields that would be the expected
// log output for a given config and any additional fields provided
func commonFields(c *Config, i *InternalFields, additional ...zap.Field) []zap.Field {
	fields := make([]zap.Field, 7)

	var b bool
	if i.IsReportable != nil {
		b = *i.IsReportable
	}

	fields[0] = String("service", c.ServiceName).field
	fields[1] = String("endpoint", i.Endpoint).field
	fields[2] = Bool("isReportable", b).field
	fields[3] = String("traceabilityID", i.TraceabilityID).field
	fields[4] = String("correlationID", i.CorrelationID).field
	fields[5] = String("userID", i.UserID).field
	fields[6] = String("clientID", i.ClientID).field

	return append(fields, additional...)
}
