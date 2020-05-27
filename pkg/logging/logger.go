// Package logging provides functionality to log into AWS Kinesis.
package logging

import (
	"github.com/caring/go-packages/pkg/logging/internal/exit"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger provides fast, structured, type safe leveled logging. All log output methods are safe for concurrent use
type Logger struct {
	// The name of the service
	serviceName string
	// a correlation ID is used to track a single user request through a
	// network of microservices.
	correlationID  string
	traceabilityID string
	clientID       string
	userID         string
	endpoint       string
	fields         []Field
	isReportable   bool
	internalLogger *zap.Logger
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

	l := Logger{
		serviceName: c.ServiceName,
		fields:      []Field{},
	}

	if *c.EnableDevLogging {
		zapConfig = newZapDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	zapConfig.Encoding = "json"
	zapConfig.Level.SetLevel(zapcore.Level(c.LogLevel))
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

// NewNopLogger returns a new logger that wont log outputs, wont error, and wont call any internal hooks
func NewNopLogger() *Logger {
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

// FieldOpts wraps internal field values that can be updated when spawning a child logger.
type FieldOpts struct {
	Endpoint       string
	CorrelationID  string
	TraceabilityID string
	ClientID       string
	UserID         string
	IsReportable   ReportFlag
	// If set to true, the existing accumulated fields will be
	// replaced with the fields passed in, a nil value writes the
	// accumulated fields to an empty value
	OverwriteAccumulatedFields bool
	// Reset values mean the field will be set to its 0 value,
	// regardless of what is passed into the opts object
	ResetEndpoint       bool
	ResetCorrelationID  bool
	ResetTraceabilityID bool
	ResetClientID       bool
	ResetUserID         bool
}

// NewChild clones logger and returns a child instance where any internal fields are overwritten
// with any non 0 values passed in, or if the field reset is set to true then the field will
// be set to a zero value. If nil options are passed in then the logger is simply cloned without change.
func (l *Logger) NewChild(opts *FieldOpts, fields ...Field) *Logger {
	new := *l
	new.with(opts, fields...)

	return &new
}

// With sets the internal fields with the provided options.
// See the options struct for more details
func (l *Logger) With(opts *FieldOpts, fields ...Field) *Logger {
	return l.with(opts, fields...)
}

func (l *Logger) with(opts *FieldOpts, fields ...Field) *Logger {
	if opts == nil {
		l.fields = append(l.fields, fields...)
		return l
	}

	if opts.ResetEndpoint {
		l.endpoint = ""
	} else if opts.Endpoint != "" {
		l.endpoint = opts.Endpoint
	}

	if opts.ResetCorrelationID {
		l.correlationID = ""
	} else if opts.CorrelationID != "" {
		l.correlationID = opts.CorrelationID
	}

	if opts.ResetTraceabilityID {
		l.traceabilityID = ""
	} else if opts.TraceabilityID != "" {
		l.traceabilityID = opts.TraceabilityID
	}

	if opts.ResetClientID {
		l.clientID = ""
	} else if opts.ClientID != "" {
		l.clientID = opts.ClientID
	}

	if opts.ResetUserID {
		l.userID = ""
	} else if opts.UserID != "" {
		l.userID = opts.UserID
	}

	if opts.IsReportable != nil {
		l.isReportable = *opts.IsReportable
	}

	if opts.OverwriteAccumulatedFields {
		l.writeFields(fields...)
	} else {
		l.accumulateFields(fields...)
	}

	return l
}

// accumulates the given fields onto the existing accumulated fields of logger
func (l *Logger) accumulateFields(f ...Field) {
	l.fields = append(l.fields, f...)
}

// overwrites the accumulated fields of logger with the fields passed in,
// a nil argument writes an empty slice to the fields
func (l *Logger) writeFields(f ...Field) {
	if f == nil {
		l.fields = []Field{}
	}
	l.fields = f
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

// DPanic logs a message at DPanicLevel. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func (l *Logger) DPanic(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.DPanic(message, f...)
}

// Fatal logs the message at fatal level output, then calls os.Exit. This includes the additional fields provided,
// the standard fields and any fields accumulated on the logger.
func (l *Logger) Fatal(message string, additionalFields ...Field) {
	// This one method differs so that we may abstract away os.Exit into a mockable
	// and testable internal library of our own. Zap has done this, but it is internal
	// so we cant use it
	f := l.getZapFields(additionalFields...)
	if ce := l.internalLogger.Check(zapcore.FatalLevel, message); ce != nil {
		ce.Should(ce.Entry, zapcore.WriteThenNoop)
		ce.Write(f...)
	}
	exit.Exit()
}

// getZapFields aggregates the Logger fields into a typed and structured set of zap fields.
func (l *Logger) getZapFields(fields ...Field) []zap.Field {
	// 7 is the number of internal fields that appear on every log entry
	total := 7 + len(fields) + len(l.fields)

	zapped := make([]zap.Field, total)

	zapped[0] = String("service", l.serviceName).field
	zapped[1] = String("endpoint", l.endpoint).field
	zapped[2] = Bool("isReportable", l.isReportable).field
	zapped[3] = String("traceabilityID", l.traceabilityID).field
	zapped[4] = String("correlationID", l.correlationID).field
	zapped[5] = String("userID", l.userID).field
	zapped[6] = String("clientID", l.clientID).field

	i := 7
	for _, f := range l.fields {
		zapped[i] = f.field
		i++
	}

	for _, f := range fields {
		zapped[i] = f.field
		i++
	}

	return zapped
}
