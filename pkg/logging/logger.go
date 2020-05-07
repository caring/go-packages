// Package logging provides functionality to log into AWS Kinesis.
package logging

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

// Logger provides debug, info, warn, panic, & fatal functions to log.
type Logger interface {
	GetInternalLogger() *zap.Logger
	NewJaegerLogger() jaeger.Logger
	NewGRPCUnaryServerInterceptor() grpc.UnaryServerInterceptor
	NewGRPCStreamServerInterceptor() grpc.StreamServerInterceptor
	NewChild(*ChildConfig, ...Field) Logger
	SetEndpoint(endpoint string) Logger
	SetServiceID(serviceID string) Logger
	SetCorrelationID(correlationID string) Logger
	SetClientID(clientID string) Logger
	SetTraceabilityID(traceabilityID string) Logger
	SetIsReportable(isReportable bool) Logger
	SetAdditionalFields(additionalFields ...Field) Logger
	AppendAdditionalFields(additionalFields ...Field) Logger
	Debug(message string, additionalFields ...Field)
	Info(message string, additionalFields ...Field)
	Warn(message string, additionalFields ...Field)
	Fatal(message string, additionalFields ...Field)
	Error(message string, additionalFields ...Field)
	Panic(message string, additionalFields ...Field)
}

type loggerImpl struct {
	writeToKinesis   bool
	serviceID        string
	correlationalID  string
	traceabilityID   string
	clientID         string
	userID           string
	endpoint         string
	additionalFields []Field
	isReportable     bool
	parentLogger     Logger
	internalLogger   *zap.Logger
}

// NewLogger initializes a new logger.
// Connects into AWS and sets up a kinesis service.
// It returns a new Logger instance that can be used as the initial parent for all application logging.
func NewLogger(
	isProd bool,
	loggerName,
	streamName,
	serviceID,
	awsAccessKey,
	awsSecretKey,
	awsRegion string,
	isAsync,
	writeToKinesis bool) (Logger, error) {
	l := loggerImpl{
		serviceID:        serviceID,
		additionalFields: nil,
		isReportable:     false,
		parentLogger:     nil,
		internalLogger:   nil,
	}

	var logger *zap.Logger

	if isProd {
		config := zap.NewProductionConfig()
		config.Encoding = "json"
		logger, _ = config.Build(zap.AddCallerSkip(1))
	} else {
		config := zap.NewDevelopmentConfig()
		config.Encoding = "json"
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stdout"}
		// This displays log messages in a format compatable with the zap-pretty print library
		config.EncoderConfig = zapcore.EncoderConfig{
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
		// Remove the wrapper from the caller display so we know which file called _our_ logger
		logger, _ = config.Build(zap.AddCallerSkip(1))
	}

	if writeToKinesis {
		cred := credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, "")
		cfg := aws.NewConfig().WithRegion(awsRegion).WithCredentials(cred)

		kcHookConstructor, err := newKinesisHook(streamName, cfg, isProd, isAsync)

		if kcHookConstructor == nil {
			return nil, err
		}

		kcHook, err := kcHookConstructor.getHook()

		if err != nil {
			return nil, err
		}

		logger = logger.WithOptions(zap.Hooks(kcHook))
	}

	logger = logger.Named(loggerName)
	l.internalLogger = logger

	return &l, nil
}

// GetInternalLogger returns the zap internal logger pointer
func (l *loggerImpl) GetInternalLogger() *zap.Logger {
	return l.internalLogger
}

// ChildConfig wraps internal field values that can be updated when spawning a child logger.
type ChildConfig struct {
	Endpoint       string
	CorrelationID  string
	TraceabilityID string
	ClientID       string
	UserID         string
	IsReportable   *bool
}

// NewChild clones logger and returns a child instance where any internal fields are overwritten
// with any non 0 values passed in
func (l *loggerImpl) NewChild(c *ChildConfig, additionalFields ...Field) Logger {
	new := *l
	new.parentLogger = l

	if c.Endpoint != "" {
		new.endpoint = c.Endpoint
	}

	if c.CorrelationID != "" {
		new.correlationalID = c.CorrelationID
	}

	if c.TraceabilityID != "" {
		new.traceabilityID = c.TraceabilityID
	}

	if c.ClientID != "" {
		new.clientID = c.ClientID
	}

	if c.UserID != "" {
		new.userID = c.UserID
	}

	if c.IsReportable != nil {
		new.isReportable = *c.IsReportable
	}

	if additionalFields != nil {
		new.additionalFields = additionalFields
	}

	return &new
}

// SetEndpoint sets the endpoint string to the existing Logger instance.
func (l *loggerImpl) SetEndpoint(endpoint string) Logger {
	l.endpoint = endpoint

	return l
}

// SetServiceID sets the serviceID string to the existing Logger instance.
func (l *loggerImpl) SetServiceID(serviceID string) Logger {
	l.serviceID = serviceID

	return l
}

// SetCorrelationID sets the string to the Logger instance.
func (l *loggerImpl) SetCorrelationID(correlationID string) Logger {
	l.correlationalID = correlationID

	return l
}

// SetClientID sets the string to the Logger instance.
func (l *loggerImpl) SetClientID(clientID string) Logger {
	l.clientID = clientID

	return l
}

// SetTraceabilityID sets the string to the Logger instance.
func (l *loggerImpl) SetTraceabilityID(traceabilityID string) Logger {
	l.traceabilityID = traceabilityID

	return l
}

// SetUserID sets the string userID to the Logger instance.
func (l *loggerImpl) SetUserID(userID string) Logger {
	l.userID = userID

	return l
}

// SetIsReportable sets the boolean isReportable to the Logger instance.
func (l *loggerImpl) SetIsReportable(isReportable bool) Logger {
	l.isReportable = isReportable

	return l
}

// SetAdditionalData sets the additionalFields map to the Logger instance.
func (l *loggerImpl) SetAdditionalFields(additionalFields ...Field) Logger {
	l.additionalFields = additionalFields

	return l
}

// AppendAdditionalFields appends Logger additionalFields map with new fields.
func (l *loggerImpl) AppendAdditionalFields(additionalFields ...Field) Logger {
	if l.additionalFields == nil {
		l.additionalFields = additionalFields
	} else if additionalFields != nil {
		l.additionalFields = append(l.additionalFields, additionalFields...)
	}

	return l
}

// Debug provides developer ability to send useful debug related  messages into Kinesis logging stream.
func (l *loggerImpl) Debug(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Debug(message, f...)
}

// Info provides developer ability to send general info  messages into Kinesis logging stream.
func (l *loggerImpl) Info(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Info(message, f...)
}

// Warn provides developer ability to send useful warning messages into Kinesis logging stream.
func (l *loggerImpl) Warn(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Warn(message, f...)
}

// Fatal provides developer ability to send application fatal messages into Kinesis logging stream.
func (l *loggerImpl) Fatal(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Fatal(message, f...)
}

// Error provides developer ability to send error  messages into Kinesis logging stream.
func (l *loggerImpl) Error(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Error(message, f...)
}

// Panic provides developer ability to send panic  messages into Kinesis logging stream.
func (l *loggerImpl) Panic(message string, additionalFields ...Field) {
	f := l.getZapFields(additionalFields...)
	l.internalLogger.Panic(message, f...)
}

// getZapFields aggregates the LogDetails and Logger into a combined map.
// It returns a json string to insert into an actual log.
func (l *loggerImpl) getZapFields(additionalFields ...Field) []zap.Field {
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

	fields[1] = NewStringField("serviceID", l.serviceID).field
	fields[0] = NewStringField("endpoint", l.endpoint).field
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
