// Package logging provides functionality to log into AWS Kinesis.
package logging

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	jaeger_zap "github.com/uber/jaeger-client-go/log/zap"
	"google.golang.org/grpc"
)


// Logger provides debug, info, warn, panic, & fatal functions to log.
type Logger struct {
	internalLogger *zap.Logger
	writeToKinesis bool
}


// LogDetails is the schema structure for logging.
type LogDetails struct {
	serviceID string
	correlationalID int64
	traceabilityID int64
	clientID int64
	userID int64
	endpoint string
	additionalData map[string]Field
	isReportable bool
	parentDetails *LogDetails
	logger *Logger
}

type TracerLogger struct {
	InfoF	func(message string, args ...interface{})
	Error	func(message string)
}


// KinesisHook provides the details to hook into the Zap logger
type KinesisHook struct {
	svc 			*kinesis.Kinesis
	Async			bool
	AcceptedLevels	[]zapcore.Level
	streamName 		string
	m				sync.Mutex
	isProd			bool
	serviceID		string
}

type Field struct {
	field zap.Field
}


// InitLogger initializes a new logger.
// Connects into AWS and sets up a kinesis service.
// It returns a new logger instance and any errors upon initialization.
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
		serviceID:       serviceID,
		additionalData:  nil,
		isReportable:    false,
		parentDetails:   nil,
		logger:          nil,
	}

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

	if isProd {
		config := zap.NewProductionConfig()
		config.Encoding = "json"
		logger, _ := config.Build()

		if writeToKinesis {
			logger = logger.WithOptions(zap.Hooks(kcHook))
		}

		logger = logger.Named(loggerName)
		l.internalLogger = logger
	} else {
		config := zap.NewDevelopmentConfig()
		config.Encoding = "json"
		logger, _ := config.Build()

		if writeToKinesis {
			logger = logger.WithOptions(zap.Hooks(kcHook))
		}

		logger = logger.Named(loggerName)
		l.internalLogger = logger
	}

	ld.logger = &l

	return ld, nil
}

func NewStringField(k, v string) Field {
	f := Field{}
	s := zap.String(k, v)
	f.field = s

	return f
}

func NewInt64Field(k string, v int64) Field {
	f := Field{}
	i := zap.Int64(k, v)
	f.field = i

	return f
}

func NewBoolField(k string, v bool) Field {
	f := Field{}
	b := zap.Bool(k, v)
	f.field = b

	return f
}


// GetInternalLogger returns the zap internal logger pointer
func (l *Logger) GetInternalLogger() *zap.Logger {
	return l.internalLogger
}


func (d *LogDetails) NewChild(
	serviceID,
	endpoint string,
	correlationalID, traceabilityID, clientID, userID int64,
	isReportable bool,
	additionalData map[string]Field) *LogDetails {
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

func (d *LogDetails) SetEndpoint(endpoint string) *LogDetails {
	d.endpoint = endpoint

	return d
}

func (d *LogDetails) SetServiceID(serviceID string) *LogDetails {
	d.serviceID = serviceID

	return d
}


func (d *LogDetails) SetCorrelationalID(correlationalID int64) *LogDetails {
	d.correlationalID = correlationalID

	return d
}

func (d *LogDetails) SetClientID(clientID int64) *LogDetails {
	d.clientID = clientID

	return d
}

func (d *LogDetails) SetTraceabilityID(traceabilityID int64) *LogDetails {
	d.traceabilityID = traceabilityID

	return d
}

func (d *LogDetails) SetUserID(userID int64) *LogDetails {
	d.userID = userID

	return d
}

func (d *LogDetails) SetIsReportable(isReportable bool) *LogDetails {
	d.isReportable = isReportable

	return d
}


func (d *LogDetails) SetAdditionalData(additionalData map[string]Field) *LogDetails {
	d.additionalData = additionalData

	return d
}

// NewTracerLogger creates a logger that implements the jaeger logger interface
// and is populated by both the loggers parent fields and the log details provided
func (d *LogDetails) NewTracerLogger() *TracerLogger {
	populatedL := d.logger.internalLogger.With(d.getLogContent(nil)...)
	l := jaeger_zap.NewLogger(populatedL)
	tl := TracerLogger{}
	tl.Error = l.Error
	tl.InfoF = l.Infof

	return &tl
}

// NewGRPCUnaryServerInterceptor creates a gRPC unary interceptor that is wrapped around
// the internal logger populated with its parents fields and any provided log details
func (d *LogDetails) NewGRPCUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	populatedL := d.logger.internalLogger.With(d.getLogContent(nil)...)

	return grpc_zap.UnaryServerInterceptor(populatedL)
}

// NewGRPCStreamServerInterceptor creates a gRPC stream interceptor that is wrapped around
// the internal logger populated with its parents fields and any provided log details
func (d *LogDetails) NewGRPCStreamServerInterceptor() grpc.StreamServerInterceptor {
	populatedL := d.logger.internalLogger.With(d.getLogContent(nil)...)

	return grpc_zap.StreamServerInterceptor(populatedL)
}


// Debug provides developer ability to send useful debug related  messages into Kinesis logging stream.
func (d *LogDetails) Debug(message string, additionalData map[string]Field) {
	c := d.getLogContent(additionalData)
	d.logger.GetInternalLogger().Debug(message, c...)
}


// Info provides developer ability to send general info  messages into Kinesis logging stream.
func (d *LogDetails) Info(message string, additionalData map[string]Field) {
	c := d.getLogContent(additionalData)
	d.logger.GetInternalLogger().Info(message, c...)
}


// Warn provides developer ability to send useful warning messages into Kinesis logging stream.
func (d *LogDetails) Warn(message string, additionalData map[string]Field) {
	c := d.getLogContent(additionalData)
	d.logger.GetInternalLogger().Warn(message, c...)
}


// Fatal provides developer ability to send application fatal messages into Kinesis logging stream.
func (d *LogDetails) Fatal(message string, additionalData map[string]Field) {
	c := d.getLogContent(additionalData)
	d.logger.GetInternalLogger().Fatal(message, c...)
}


// Error provides developer ability to send error  messages into Kinesis logging stream.
func (d *LogDetails) Error(message string, additionalData map[string]Field) {
	c := d.getLogContent(additionalData)
	d.logger.GetInternalLogger().Error(message, c...)
}


// Panic provides developer ability to send panic  messages into Kinesis logging stream.
func (d *LogDetails) Panic(message string, additionalData map[string]Field) {
	c := d.getLogContent(additionalData)
	d.logger.GetInternalLogger().Panic(message, c...)
}

// getLogContents aggregates the LogDetails and Logger into a combined map.
// It returns a json string to insert into an actual log.
func (d *LogDetails) getLogContent(additionalData map[string]Field) []zap.Field {
	if d.additionalData == nil {
		d.additionalData = additionalData
	} else if additionalData != nil {
		for k, v := range additionalData {
			d.additionalData[k] = v
		}
	}

	sliceTotal := 7

	if d.additionalData != nil && len(d.additionalData) > 0 {
		sliceTotal = sliceTotal + len(d.additionalData)
	}

	fields := make([]zap.Field, sliceTotal)

	fields[0] = NewStringField("endpoint", d.endpoint).field
	fields[1] = NewStringField("serviceID", d.serviceID).field
	fields[2] = NewBoolField("isReportable", d.isReportable).field
	fields[3] = NewInt64Field("traceabilityID", d.traceabilityID).field
	fields[4] = NewInt64Field("correlationalID", d.correlationalID).field
	fields[5] = NewInt64Field("userID", d.userID).field
	fields[6] = NewInt64Field("clientID", d.clientID).field


	if d.additionalData != nil {
		ind := 7
		for _, fieldData := range d.additionalData {
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
		streamName:		streamName,
		svc:            kc,
		AcceptedLevels: AllLevels,
		m:              sync.Mutex{},
		isProd:			isProd,
		Async: 			isAsync,
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
				Data:                      []byte(e.Message),
				StreamName:                aws.String(ch.streamName),
				PartitionKey: 			   &partKey,
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

			if !ch.isProd  && len(records.Records) > 0 {
				lastRecord := len(records.Records)  -1
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
