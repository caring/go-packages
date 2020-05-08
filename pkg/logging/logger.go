// Package logging provides functionality to log into AWS Kinesis.
package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger provides fast, structured, type safe leveled logging. A logger instance
// wraps the standard caring log structure and has methods for setting each of these values.
// There are also utils for obtaining middleware that wraps loggers for our common stack pieces
type Logger struct {
	writeToKinesis   bool
	serviceName      string
	correlationalID  string
	traceabilityID   string
	clientID         string
	userID           string
	endpoint         string
	additionalFields []Field
	isReportable     bool
	internalLogger   *zap.Logger
}

// NewLogger initializes a new logger.
// Connects into AWS and sets up a kinesis service.
func NewLogger(config *Config) (*Logger, error) {
	var (
		zapConfig zap.Config
	)

	c, err := mergeAndPopulateConfig(config)
	if err != nil {
		return nil, err
	}

	l := Logger{}

	if *c.EnableDevLogging {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.OutputPaths = []string{"stdout"}
		zapConfig.ErrorOutputPaths = []string{"stdout"}
		// This displays log messages in a format compatable with the zap-pretty print library
		zapConfig.EncoderConfig = zapcore.EncoderConfig{
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
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	zapConfig.Encoding = "json"
	zapConfig.Level.UnmarshalText([]byte(c.LogLevel))
	// caller skip makes the caller appear as the line of code where this package is called,
	// instead of where zap is called in this package
	zapL, err := zapConfig.Build(zap.AddCallerSkip(1))

	if err != nil {
		return nil, err
	}

	if !*c.DisableKinesis {
		kcHookConstructor, err := newKinesisHook(c.KinesisStreamName, c.KinesisPartitionKey)

		if kcHookConstructor == nil {
			return nil, err
		}

		kcHook, err := kcHookConstructor.getHook()

		if err != nil {
			return nil, err
		}

		zapL = zapL.WithOptions(zap.Hooks(kcHook))
	}

	zapL = zapL.Named(c.LoggerName)
	l.internalLogger = zapL

	return &l, nil
}

// NewNoOpLogger returns a new logger that wont log outputs, wont error, and wont call any internal hooks
func NewNoOpLogger() *Logger {
	return &Logger{
		internalLogger: zap.NewNop(),
	}
}

// GetInternalLogger returns the zap internal logger pointer.
// Note: Zap should not be considered a stable dependency, another logger
// may be substituted at any time
func (l *Logger) GetInternalLogger() *zap.Logger {
	return l.internalLogger
}

// InternalFields wraps internal field values that can be updated when spawning a child logger.
type InternalFields struct {
	Endpoint       string
	CorrelationID  string
	TraceabilityID string
	ClientID       string
	UserID         string
	IsReportable   *bool
}

// NewChild clones logger and returns a child instance where any internal fields are overwritten
// with any non 0 values passed in
func (l *Logger) NewChild(i *InternalFields, additionalFields ...Field) *Logger {
	new := *l

	if i.Endpoint != "" {
		new.endpoint = i.Endpoint
	}

	if i.CorrelationID != "" {
		new.correlationalID = i.CorrelationID
	}

	if i.TraceabilityID != "" {
		new.traceabilityID = i.TraceabilityID
	}

	if i.ClientID != "" {
		new.clientID = i.ClientID
	}

	if i.UserID != "" {
		new.userID = i.UserID
	}

	if i.IsReportable != nil {
		new.isReportable = *i.IsReportable
	}

	if additionalFields != nil {
		new.additionalFields = additionalFields
	}

	return &new
}

// SetEndpoint sets the endpoint string to the existing Logger instance.
func (l *Logger) SetEndpoint(endpoint string) *Logger {
	l.endpoint = endpoint

	return l
}

// SetServiceName sets the serviceName string to the existing Logger instance.
func (l *Logger) SetServiceName(serviceName string) *Logger {
	l.serviceName = serviceName

	return l
}

// SetCorrelationID sets the string to the Logger instance.
func (l *Logger) SetCorrelationID(correlationID string) *Logger {
	l.correlationalID = correlationID

	return l
}

// SetClientID sets the string to the Logger instance.
func (l *Logger) SetClientID(clientID string) *Logger {
	l.clientID = clientID

	return l
}

// SetTraceabilityID sets the string to the Logger instance.
func (l *Logger) SetTraceabilityID(traceabilityID string) *Logger {
	l.traceabilityID = traceabilityID

	return l
}

// SetUserID sets the string userID to the Logger instance.
func (l *Logger) SetUserID(userID string) *Logger {
	l.userID = userID

	return l
}

// SetIsReportable sets the boolean isReportable to the Logger instance.
func (l *Logger) SetIsReportable(isReportable bool) *Logger {
	l.isReportable = isReportable

	return l
}

// SetAdditionalFields sets the additionalFields map to the Logger instance.
func (l *Logger) SetAdditionalFields(additionalFields ...Field) *Logger {
	l.additionalFields = additionalFields

	return l
}

// AppendAdditionalFields appends Logger additionalFields map with new fields.
func (l *Logger) AppendAdditionalFields(additionalFields ...Field) *Logger {
	if l.additionalFields == nil {
		l.additionalFields = additionalFields
	} else if additionalFields != nil {
		l.additionalFields = append(l.additionalFields, additionalFields...)
	}

	return l
}

// Debug provides developer ability to send useful debug related  messages into Kinesis logging stream.
func (l *Logger) Debug(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Debug(message, f...)
}

// Info provides developer ability to send general info  messages into Kinesis logging stream.
func (l *Logger) Info(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Info(message, f...)
}

// Warn provides developer ability to send useful warning messages into Kinesis logging stream.
func (l *Logger) Warn(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Warn(message, f...)
}

// Fatal provides developer ability to send application fatal messages into Kinesis logging stream.
func (l *Logger) Fatal(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Fatal(message, f...)
}

// Error provides developer ability to send error  messages into Kinesis logging stream.
func (l *Logger) Error(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Error(message, f...)
}

// Panic provides developer ability to send panic  messages into Kinesis logging stream.
func (l *Logger) Panic(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Panic(message, f...)
}

// getZapFields aggregates the LogDetails and Logger into a combined map.
// It returns a json string to insert into an actual log.
func (l *Logger) getZapFields(additionalFields ...Field) []zap.Field {
	ad := l.additionalFields
	if ad == nil {
		ad = additionalFields
	} else if len(additionalFields) > 0 {
		ad = append(ad, additionalFields...)
	}

	sliceTotal := 7

	if ad != nil && len(ad) > 0 {
		sliceTotal = sliceTotal + len(ad)
	}

	fields := make([]zap.Field, sliceTotal)

	fields[0] = NewStringField("serviceName", l.serviceName).field
	fields[1] = NewStringField("endpoint", l.endpoint).field
	fields[2] = NewBoolField("isReportable", l.isReportable).field
	fields[3] = NewStringField("traceabilityID", l.traceabilityID).field
	fields[4] = NewStringField("correlationID", l.correlationalID).field
	fields[5] = NewStringField("userID", l.userID).field
	fields[6] = NewStringField("clientID", l.clientID).field

	if len(ad) > 0 {
		ind := 7
		for _, fieldData := range ad {
			fields[ind] = fieldData.field
			ind++
		}
	}

	return fields
}
