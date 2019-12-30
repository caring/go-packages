// Package logging provides functionality to log into AWS Kinesis.
package logging

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	jaeger_zap "github.com/uber/jaeger-client-go/log/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

// LogDetails is the schema structure for logging.
type LogDetails struct {
	Message string
	// I don't actually know that these will be int64, just a guess to demonstrate an idea
	CorrelationalID int64
	TraceabilityID  int64
	ClientID        int64
	UserID          int64
	Endpoint        string
	IsReportable    bool
	AdditionalData  map[string]interface{}
}

// Logger provides debug, info, warn, panic, & fatal functions to log.
type Logger struct {
	internalLogger *zap.Logger
	serviceID      string
	writeToKinesis bool
	parentDetails  LogDetails
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

// InitLogger initializes a new logger.
// Connects into AWS and sets up a kinesis service.
// It returns a new logger instance and any errors upon initialization.
func InitLogger(isProd bool, loggerName, streamName, serviceID, awsAccessKey, awsSecretKey, awsRegion string, isAsync, writeToKinesis bool) (Logger, error) {
	l := Logger{}
	l.serviceID = serviceID

	cred := credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, "")
	cfg := aws.NewConfig().WithRegion(awsRegion).WithCredentials(cred)

	kcHookConstructor, err := newKinesisHook(streamName, cfg, isProd, isAsync)

	if kcHookConstructor == nil {
		return l, err
	}

	kcHook, err := kcHookConstructor.getHook()

	if err != nil {
		return l, err
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

	return l, nil
}

// NewJaegerLogger creates a logger that implements the jaeger logger interface
// and is populated by both the loggers parent fields and the log details provided
func (l *Logger) NewJaegerLogger(ld LogDetails) *jaeger_zap.Logger {
	ld = l.mergeLogDetails(ld)
	populatedL := l.internalLogger.With(l.getLogContent(ld)...)

	return jaeger_zap.NewLogger(populatedL)
}

// NewGRPCUnaryServerInterceptor creates a gRPC unary interceptor that is wrapped around
// the internal logger populated with its parents fields and any provided log details
func (l *Logger) NewGRPCUnaryServerInterceptor(ld LogDetails) grpc.UnaryServerInterceptor {
	ld = l.mergeLogDetails(ld)
	populatedL := l.internalLogger.With(l.getLogContent(ld)...)

	return grpc_zap.UnaryServerInterceptor(populatedL)
}

// NewGRPCStreamServerInterceptor creates a gRPC stream interceptor that is wrapped around
// the internal logger populated with its parents fields and any provided log details
func (l *Logger) NewGRPCStreamServerInterceptor(ld LogDetails) grpc.StreamServerInterceptor {
	ld = l.mergeLogDetails(ld)
	populatedL := l.internalLogger.With(l.getLogContent(ld)...)

	return grpc_zap.StreamServerInterceptor(populatedL)
}

// GetInternalLogger returns the zap internal logger pointer
func (l *Logger) GetInternalLogger() *zap.Logger {
	return l.internalLogger
}

// SpawnChildLogger creates a new child logger that contains all of the fields already populated in its parent
// plus any changes specified in the provided LogDetails . changes to the child do not affect the parrent and vice-versa
func (l *Logger) SpawnChildLogger(ld LogDetails) *Logger {
	newL := Logger(*l)

	newL.parentDetails = l.mergeLogDetails(ld)

	return &newL
}

// Debug provides developer ability to send useful debug related  messages into Kinesis logging stream.
func (l *Logger) Debug(ld LogDetails) {
	ld = l.mergeLogDetails(ld)
	l.internalLogger.Debug(ld.Message, l.getLogContent(ld)...)
}

// Info provides developer ability to send general info  messages into Kinesis logging stream.
func (l *Logger) Info(ld LogDetails) {
	ld = l.mergeLogDetails(ld)
	l.internalLogger.Info(ld.Message, l.getLogContent(ld)...)
}

// Warn provides developer ability to send useful warning messages into Kinesis logging stream.
func (l *Logger) Warn(ld LogDetails) {
	ld = l.mergeLogDetails(ld)
	l.internalLogger.Warn(ld.Message, l.getLogContent(ld)...)
}

// Fatal provides developer ability to send application fatal messages into Kinesis logging stream.
func (l *Logger) Fatal(ld LogDetails) {
	ld = l.mergeLogDetails(ld)
	l.internalLogger.Fatal(ld.Message, l.getLogContent(ld)...)
}

// Error provides developer ability to send error  messages into Kinesis logging stream.
func (l *Logger) Error(ld LogDetails) {
	ld = l.mergeLogDetails(ld)
	l.internalLogger.Error(ld.Message, l.getLogContent(ld)...)
}

// Panic provides developer ability to send panic  messages into Kinesis logging stream.
func (l *Logger) Panic(ld LogDetails) {
	ld = l.mergeLogDetails(ld)
	l.internalLogger.Panic(ld.Message, l.getLogContent(ld)...)
}

// getLogContents aggregates the LogDetails and Logger a slice
// of typed zap fields
func (l *Logger) getLogContent(ld LogDetails) []zap.Field {
	// We know the log details types so avoid the overhead of marshaling and the errors that come with it.
	// Passing fields along instead of marshalled JSON also makes all the fields top level instead of all nested under
	// the logs built in message field
	s := []zap.Field{
		zap.Int64("userId", ld.UserID),
		zap.Int64("traceabilityId", ld.TraceabilityID),
		zap.String("endpoint", ld.Endpoint),
		zap.Int64("correlationalId", ld.CorrelationalID),
		zap.Int64("clientId", ld.ClientID),
		zap.String("serviceId", l.serviceID),
		zap.Bool("isReportable", ld.IsReportable),
	}

	for k, v := range ld.AdditionalData {
		s = append(s, getZapType(k, v))
	}

	return s
}

// Takes input log details and backfills any empty fields with parent details in the logger
func (l *Logger) mergeLogDetails(ld LogDetails) LogDetails {
	if ld.Message == "" {
		ld.Message = l.parentDetails.Message
	}
	if ld.CorrelationalID == 0 {
		ld.CorrelationalID = l.parentDetails.CorrelationalID
	}
	if ld.TraceabilityID == 0 {
		ld.TraceabilityID = l.parentDetails.TraceabilityID
	}
	if ld.ClientID == 0 {
		ld.ClientID = l.parentDetails.ClientID
	}
	if ld.UserID == 0 {
		ld.UserID = l.parentDetails.UserID
	}
	if ld.Endpoint == "" {
		ld.Endpoint = l.parentDetails.Endpoint
	}
	if ld.IsReportable == false {
		ld.IsReportable = l.parentDetails.IsReportable
	}
	if len(ld.AdditionalData) == 0 {
		ld.AdditionalData = l.parentDetails.AdditionalData
	}
	return ld
}

func getZapType(k string, v interface{}) zap.Field {
	switch v.(type) {
	case string:
		s, ok := v.(string)
		if !ok {
			// swallow errors and return a placeholder field instead so
			// an error doesn't cause an entire log write of data to be lost in the warehouse
			return zap.String(fmt.Sprint("ErrField", k), "FIELD_TYPE_CONVERSION_ERROR")
		}
		return zap.String(k, s)
	case int64:
		i, ok := v.(int64)
		if !ok {
			return zap.String(fmt.Sprint("ErrField", k), "FIELD_TYPE_CONVERSION_ERROR")
		}
		return zap.Int64(k, i)
	case bool:
		b, ok := v.(bool)
		if !ok {
			return zap.String(fmt.Sprint("ErrField", k), "FIELD_TYPE_CONVERSION_ERROR")
		}
		return zap.Bool(k, b)
	case float64:
		f, ok := v.(float64)
		if !ok {
			return zap.String(fmt.Sprint("ErrField", k), "FIELD_TYPE_CONVERSION_ERROR")
		}
		return zap.Float64(k, f)
	case map[string]interface{}:
		j, err := json.Marshal(v)
		if err != nil {
			return zap.String(fmt.Sprint("ErrField", k), "FIELD_TYPE_CONVERSION_ERROR")
		}
		return zap.String(k, string(j))
	default:
		return zap.String(fmt.Sprint("ErrField", k), "UNHANDLED_FIELD_TYPE")
	}
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
