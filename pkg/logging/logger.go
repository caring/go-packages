// Package logging provides functionality to log into AWS Kinesis.
package logging

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/uber/jaeger-client-go"
	jaeger_zap "github.com/uber/jaeger-client-go/log/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

// Logger provides debug, info, warn, panic, & fatal functions to log.
type Logger struct {
	internalLogger *zap.Logger
	writeToKinesis bool
}

// LogDetails is the schema structure for logging.
type LogDetails struct {
	serviceID       string
	correlationalID int64
	traceabilityID  int64
	clientID        int64
	userID          int64
	endpoint        string
	additionalData  []Field
	isReportable    bool
	parentDetails   *LogDetails
	logger          *Logger
}

// KinesisHook provides the details to hook into the Zap logger
type KinesisHook struct {
	svc            *kinesis.Kinesis
	Async          bool
	AcceptedLevels []zapcore.Level
	streamName     string
	m              sync.Mutex
	isProd         bool
	serviceID      string
}

// Field provides a wrapping struct to be used internally to log indexable fields.
type Field struct {
	field zap.Field
}

// InitLogging initializes a new logger.
// Connects into AWS and sets up a kinesis service.
// It returns a new LogDetails instance that can be used as the initial parent for all application logging.
func InitLogging(
	isProd bool,
	loggerName,
	streamName,
	serviceID,
	awsAccessKey,
	awsSecretKey,
	awsRegion string,
	isAsync,
	writeToKinesis bool) (LogDetails, error) {
	l := Logger{}
	ld := LogDetails{
		serviceID:      serviceID,
		additionalData: nil,
		isReportable:   false,
		parentDetails:  nil,
		logger:         nil,
	}

	var logger *zap.Logger

	if isProd {
		config := zap.NewProductionConfig()
		config.Encoding = "json"
		logger, _ = config.Build()
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
			return ld, err
		}

		kcHook, err := kcHookConstructor.getHook()

		if err != nil {
			return ld, err
		}

		logger = logger.WithOptions(zap.Hooks(kcHook))
	}

	logger = logger.Named(loggerName)
	l.internalLogger = logger

	ld.logger = &l

	return ld, nil
}

// NewStringField creates an string field used for log indexing.
func NewStringField(k, v string) Field {
	f := Field{}
	s := zap.String(k, v)
	f.field = s

	return f
}

// NewInt64Field creates an int64 field used for log indexing.
func NewInt64Field(k string, v int64) Field {
	f := Field{}
	i := zap.Int64(k, v)
	f.field = i

	return f
}

// NewFloat64Field creates an float64 field used for log indexing.
func NewFloat64Field(k string, v float64) Field {
	f := Field{}
	fl := zap.Float64(k, v)
	f.field = fl

	return f
}

// NewBoolField creates a bool field used for log indexing.
func NewBoolField(k string, v bool) Field {
	f := Field{}
	b := zap.Bool(k, v)
	f.field = b

	return f
}

// NewAnyField takes a key and an arbitrary value and chooses the
// best way to represent them as a field, falling back to a reflection-based
// approach only if necessary.
func NewAnyField(k string, v interface{}) Field {
	f := Field{}
	a := zap.Any(k, v)
	f.field = a

	return f
}

// NewStringsField creates an array of strings field for log indexing
func NewStringsField(k string, vs []string) Field {
	f := Field{}
	ss := zap.Strings(k, vs)
	f.field = ss

	return f
}

// NewInt64sField creates an array of int64s field for log indexing
func NewInt64sField(k string, vs []int64) Field {
	f := Field{}
	is := zap.Int64s(k, vs)
	f.field = is

	return f
}

// NewFloat64sField creates an array of int64s field for log indexing
func NewFloat64sField(k string, vs []float64) Field {
	f := Field{}
	fs := zap.Float64s(k, vs)
	f.field = fs

	return f
}

// NewBoolsField creates an array of bools field used for log indexing.
func NewBoolsField(k string, vs []bool) Field {
	f := Field{}
	bs := zap.Bools(k, vs)
	f.field = bs

	return f
}

// GetInternalLogger returns the zap internal logger pointer
func (l *Logger) GetInternalLogger() *zap.Logger {
	return l.internalLogger
}

// NewChild creates a new child LogDetails struct based on the current.
// It clones the existing one and updates with any non-zero parameters.
// Returns a pointer reference to the new child LogDetails instance.
func (d *LogDetails) NewChild(
	serviceID,
	endpoint string,
	correlationalID, traceabilityID, clientID, userID int64,
	isReportable bool,
	additionalData []Field) *LogDetails {
	ld := d
	ld.parentDetails = d

	if serviceID != "" {
		ld.serviceID = serviceID
	}

	if correlationalID != 0 {
		ld.correlationalID = correlationalID
	}

	if traceabilityID != 0 {
		ld.traceabilityID = traceabilityID
	}

	if clientID != 0 {
		ld.clientID = clientID
	}

	if userID != 0 {
		ld.userID = userID
	}

	ld.isReportable = isReportable

	if additionalData != nil {
		ld.additionalData = additionalData
	}

	if endpoint != "" {
		ld.endpoint = endpoint
	}

	return ld
}

// SetEndpoint sets the endpoint string to the existing LogDetails instance.
func (d *LogDetails) SetEndpoint(endpoint string) *LogDetails {
	d.endpoint = endpoint

	return d
}

// SetServiceID sets the serviceID string to the existing LogDetails instance.
func (d *LogDetails) SetServiceID(serviceID string) *LogDetails {
	d.serviceID = serviceID

	return d
}

// SetCorrelationalID sets the int64 to the LogDetails instance.
func (d *LogDetails) SetCorrelationalID(correlationalID int64) *LogDetails {
	d.correlationalID = correlationalID

	return d
}

// SetClientID sets the string to the LogDetails instance.
func (d *LogDetails) SetClientID(clientID int64) *LogDetails {
	d.clientID = clientID

	return d
}

// SetTraceabilityID sets the int64 to the LogDetails instance.
func (d *LogDetails) SetTraceabilityID(traceabilityID int64) *LogDetails {
	d.traceabilityID = traceabilityID

	return d
}

// SetUserID sets the int64 userID to the LogDetails instance.
func (d *LogDetails) SetUserID(userID int64) *LogDetails {
	d.userID = userID

	return d
}

// SetIsReportable sets the boolean isReportable to the LogDetails instance.
func (d *LogDetails) SetIsReportable(isReportable bool) *LogDetails {
	d.isReportable = isReportable

	return d
}

// SetAdditionalData sets the additionalData map to the LogDetails instance.
func (d *LogDetails) SetAdditionalData(additionalData []Field) *LogDetails {
	d.additionalData = additionalData

	return d
}

// AppendAdditionalData appends LogDetails additionalData map with new fields.
func (d *LogDetails) AppendAdditionalData(additionalData []Field) *LogDetails {
	if d.additionalData == nil {
		d.additionalData = additionalData
	} else if additionalData != nil {
		d.additionalData = append(d.additionalData, additionalData...)
	}

	return d
}

// NewJaegerLogger creates a logger that implements the jaeger logger interface
// and is populated by both the loggers parent fields and the log details provided
func (d *LogDetails) NewJaegerLogger() jaeger.Logger {
	populatedL := d.logger.internalLogger.With(d.getLogContent()...)
	l := jaeger_zap.NewLogger(populatedL)

	return l
}

// NewGRPCUnaryServerInterceptor creates a gRPC unary interceptor that is wrapped around
// the internal logger populated with its parents fields and any provided log details
func (d *LogDetails) NewGRPCUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	populatedL := d.logger.internalLogger.With(d.getLogContent()...)

	return grpc_zap.UnaryServerInterceptor(populatedL)
}

// NewGRPCStreamServerInterceptor creates a gRPC stream interceptor that is wrapped around
// the internal logger populated with its parents fields and any provided log details
func (d *LogDetails) NewGRPCStreamServerInterceptor() grpc.StreamServerInterceptor {
	populatedL := d.logger.internalLogger.With(d.getLogContent()...)

	return grpc_zap.StreamServerInterceptor(populatedL)
}

// Debug provides developer ability to send useful debug related  messages into Kinesis logging stream.
func (d *LogDetails) Debug(message string, additionalData ...Field) {
	c := d.getLogContent(additionalData...)
	d.logger.GetInternalLogger().Debug(message, c...)
}

// Info provides developer ability to send general info  messages into Kinesis logging stream.
func (d *LogDetails) Info(message string, additionalData ...Field) {
	c := d.getLogContent(additionalData...)
	d.logger.GetInternalLogger().Info(message, c...)
}

// Warn provides developer ability to send useful warning messages into Kinesis logging stream.
func (d *LogDetails) Warn(message string, additionalData ...Field) {
	c := d.getLogContent(additionalData...)
	d.logger.GetInternalLogger().Warn(message, c...)
}

// Fatal provides developer ability to send application fatal messages into Kinesis logging stream.
func (d *LogDetails) Fatal(message string, additionalData ...Field) {
	c := d.getLogContent(additionalData...)
	d.logger.GetInternalLogger().Fatal(message, c...)
}

// Error provides developer ability to send error  messages into Kinesis logging stream.
func (d *LogDetails) Error(message string, additionalData ...Field) {
	c := d.getLogContent(additionalData...)
	d.logger.GetInternalLogger().Error(message, c...)
}

// Panic provides developer ability to send panic  messages into Kinesis logging stream.
func (d *LogDetails) Panic(message string, additionalData ...Field) {
	c := d.getLogContent(additionalData...)
	d.logger.GetInternalLogger().Panic(message, c...)
}

// getLogContents aggregates the LogDetails and Logger into a combined map.
// It returns a json string to insert into an actual log.
func (d *LogDetails) getLogContent(additionalData ...Field) []zap.Field {
	ad := d.additionalData
	if ad == nil {
		ad = additionalData
	} else if len(additionalData) > 0 {
		ad = append(ad, additionalData...)
	}

	sliceTotal := 7

	if ad != nil && len(ad) > 0 {
		sliceTotal = sliceTotal + len(ad)
	}

	fields := make([]zap.Field, sliceTotal)

	fields[0] = NewStringField("endpoint", d.endpoint).field
	fields[1] = NewStringField("serviceID", d.serviceID).field
	fields[2] = NewBoolField("isReportable", d.isReportable).field
	fields[3] = NewInt64Field("traceabilityID", d.traceabilityID).field
	fields[4] = NewInt64Field("correlationalID", d.correlationalID).field
	fields[5] = NewInt64Field("userID", d.userID).field
	fields[6] = NewInt64Field("clientID", d.clientID).field

	if len(ad) > 0 {
		ind := 7
		for _, fieldData := range ad {
			fields[ind] = fieldData.field
			ind++
		}
	}

	return fields
}

// newKinesisHook creates a KinesisHook struct to to use in the zap log.
// Tries to find the existing aws Kinesis stream.
// Creates stream when doesn't exist.
// Returns a pointer with a implemented KinesisHook.
func newKinesisHook(streamName string, cfg *aws.Config, isProd, isAsync bool) (*KinesisHook, error) {
	s := session.New(cfg)
	kc := kinesis.New(s)

	_, err := kc.DescribeStream(&kinesis.DescribeStreamInput{StreamName: aws.String(streamName)})

	// Create stream if doesn't exist
	if err != nil {
		_, err := kc.CreateStream(&kinesis.CreateStreamInput{
			ShardCount: aws.Int64(1),
			StreamName: aws.String(streamName),
		})

		if err != nil {
			return nil, err
		}

		if err := kc.WaitUntilStreamExists(&kinesis.DescribeStreamInput{StreamName: aws.String(streamName)}); err != nil {
			return nil, err
		}
	}

	ks := &KinesisHook{
		streamName:     streamName,
		svc:            kc,
		AcceptedLevels: AllLevels,
		m:              sync.Mutex{},
		isProd:         isProd,
		Async:          isAsync,
	}

	return ks, nil
}

// getHook inserts the function to use when zap creates a log entry.
func (ch *KinesisHook) getHook() (func(zapcore.Entry) error, error) {
	kWriter := func(e zapcore.Entry) error {
		if !ch.isAcceptedLevel(e.Level) {
			return nil
		}

		writer := func() error {
			partKey := "logging-1"

			putOutput, err := ch.svc.PutRecord(&kinesis.PutRecordInput{
				Data:         []byte(e.Message),
				StreamName:   aws.String(ch.streamName),
				PartitionKey: &partKey,
			})

			if err != nil {
				return err
			}

			// retrieve iterator
			iteratorOutput, err := ch.svc.GetShardIterator(&kinesis.GetShardIteratorInput{
				// Shard Id is provided when making put record(s) request.
				ShardId:           putOutput.ShardId,
				ShardIteratorType: aws.String("TRIM_HORIZON"),
				// ShardIteratorType: aws.String("AT_SEQUENCE_NUMBER"),
				// ShardIteratorType: aws.String("LATEST"),
				StreamName: aws.String(ch.streamName),
			})
			if err != nil {
				return err
			}

			// get records use shard iterator for making request
			records, err := ch.svc.GetRecords(&kinesis.GetRecordsInput{
				ShardIterator: iteratorOutput.ShardIterator,
			})
			if err != nil {
				return err
			}

			if !ch.isProd && len(records.Records) > 0 {
				lastRecord := len(records.Records) - 1
				println(records.Records[lastRecord].String())
			}

			return err
		}

		if ch.Async {
			go writer()

			return nil
		} else {
			return writer()
		}
	}

	return kWriter, nil
}

// Levels sets which levels to sent to kinesis
func (ch *KinesisHook) levels() []zapcore.Level {
	if ch.AcceptedLevels == nil {
		return AllLevels
	}
	return ch.AcceptedLevels
}

func (ch *KinesisHook) isAcceptedLevel(level zapcore.Level) bool {
	for _, lv := range ch.levels() {
		if lv == level {
			return true
		}
	}
	return false
}

// AllLevels Supported log levels
var AllLevels = []zapcore.Level{
	zapcore.DebugLevel,
	zapcore.InfoLevel,
	zapcore.WarnLevel,
	zapcore.ErrorLevel,
	zapcore.FatalLevel,
	zapcore.PanicLevel,
}
