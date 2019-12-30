// Package logging provides functionality to log into AWS Kinesis.
package logging

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger provides debug, info, warn, panic, & fatal functions to log as well as the schema structure for logging.
type Logger struct {
	Message string
	// I don't actually know that these will be int64, just a guess to demonstrate an idea
	CorrelationalID int64
	TraceabilityID  int64
	ClientID        int64
	UserID          int64
	Endpoint        string
	AdditionalData  map[string]interface{}
	IsReportable    bool

	// Internal
	internalLogger *zap.Logger
	serviceID      string
	writeToKinesis bool
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

// GetInternalLogger returns the zap internal logger pointer
func (l *Logger) GetInternalLogger() *zap.Logger {
	return l.internalLogger
}

// SpawnChildLogger creates a new child logger that contains all of the fields already populated in its parent
// changes to the child do not affect the parrent and vice-versa
func (l *Logger) SpawnChildLogger() *Logger {
	newL := Logger(*l)

	return &newL
}

// Debug provides developer ability to send useful debug related  messages into Kinesis logging stream.
func (l *Logger) Debug(message string) {
	if message != "" {
		l.Message = message
	}

	l.internalLogger.Debug(l.Message, l.getLogContent()...)
}

// Info provides developer ability to send general info  messages into Kinesis logging stream.
func (l *Logger) Info(message string) {
	if message != "" {
		l.Message = message
	}

	l.internalLogger.Info(l.Message, l.getLogContent()...)
}

// Warn provides developer ability to send useful warning messages into Kinesis logging stream.
func (l *Logger) Warn(message string) {
	if message != "" {
		l.Message = message
	}

	l.internalLogger.Warn(l.Message, l.getLogContent()...)
}

// Fatal provides developer ability to send application fatal messages into Kinesis logging stream.
func (l *Logger) Fatal(message string) {
	if message != "" {
		l.Message = message
	}

	l.internalLogger.Fatal(l.Message, l.getLogContent()...)
}

// Error provides developer ability to send error  messages into Kinesis logging stream.
func (l *Logger) Error(message string) {
	if message != "" {
		l.Message = message
	}

	l.internalLogger.Error(l.Message, l.getLogContent()...)
}

// Panic provides developer ability to send panic  messages into Kinesis logging stream.
func (l *Logger) Panic(message string) {
	if message != "" {
		l.Message = message
	}

	l.internalLogger.Panic(l.Message, l.getLogContent()...)
}

// getLogContents aggregates the LogDetails and Logger a slice
// of typed zap fields
func (l *Logger) getLogContent() []zap.Field {

	// We know the log details types so avoid the overhead of marshaling and the errors that come with it.
	// Passing fields along instead of marshalled JSON also makes all the fields top level instead of all nested under
	// the logs built in message field
	s := []zap.Field{
		zap.Int64("userID", l.UserID),
		zap.Int64("traceabilityID", l.TraceabilityID),
		zap.String("endpoint", l.Endpoint),
		zap.Int64("correlationalID", l.CorrelationalID),
		zap.Int64("clientID", l.ClientID),
		zap.String("serviceID", l.serviceID),
		zap.Bool("isReportable", l.IsReportable),
	}

	for k, v := range l.AdditionalData {
		s = append(s, getZapType(k, v))
	}

	return s
}

func getZapType(k string, v interface{}) zap.Field {
	switch v.(type) {
	case string:
		s, ok := v.(string)
		if !ok {
			// swallow errors and return a placeholder field instead so
			// an error doesn't take down logging
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
		// Etc... handle as many types as is necessary for our logging
		return zap.Bool(k, true)
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
