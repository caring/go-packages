// Package logging provides functionality to log into AWS Kinesis.
package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger provides fast, structured, type safe leveled logging. All log output methods are safe for concurrent use
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

// NewLogger initializes a new logger and connects it to a kinesis stream if enabled
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

// SetAdditionalFields overwrites the existing accumulated fields on the logger
func (l *Logger) SetAdditionalFields(additionalFields ...Field) *Logger {
	l.additionalFields = additionalFields

	return l
}

// AppendAdditionalFields accumulates fields onto the logger
func (l *Logger) AppendAdditionalFields(additionalFields ...Field) *Logger {
	if l.additionalFields == nil {
		l.additionalFields = additionalFields
	} else if additionalFields != nil {
		l.additionalFields = append(l.additionalFields, additionalFields...)
	}

	return l
}

// Debug logs the message at debug level output. This includes the additional fields provided,
// the standard fields and any fields accumulated on the logger.
func (l *Logger) Debug(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Debug(message, f...)
}

// Info logs the message at info level output. This includes the additional fields provided,
// the standard fields and any fields accumulated on the logger.
func (l *Logger) Info(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Info(message, f...)
}

// Warn logs the message at warn level output. This includes the additional fields provided,
// the standard fields and any fields accumulated on the logger.
func (l *Logger) Warn(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Warn(message, f...)
}

// Fatal logs the message at fatal level output. This includes the additional fields provided,
// the standard fields and any fields accumulated on the logger.
func (l *Logger) Fatal(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Fatal(message, f...)
}

// Error logs the message at error level output. This includes the additional fields provided,
// the standard fields and any fields accumulated on the logger.
func (l *Logger) Error(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Error(message, f...)
}

// Panic logs the message at panic level output, then panics. This includes the additional fields provided,
// the standard fields and any fields accumulated on the logger.
func (l *Logger) Panic(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Panic(message, f...)
}

// getZapFields aggregates the Logger fields into a typed and structured set of zap fields.
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

	fields[0] = String("service", l.serviceName).field
	fields[1] = String("endpoint", l.endpoint).field
	fields[2] = Bool("isReportable", l.isReportable).field
	fields[3] = String("traceabilityID", l.traceabilityID).field
	fields[4] = String("correlationID", l.correlationalID).field
	fields[5] = String("userID", l.userID).field
	fields[6] = String("clientID", l.clientID).field

	if len(ad) > 0 {
		ind := 7
		for _, fieldData := range ad {
			fields[ind] = fieldData.field
			ind++
		}
	}

	return fields
}
