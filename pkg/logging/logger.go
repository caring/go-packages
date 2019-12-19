// Package logging provides functionality to log into AWS Kinesis.
package logging

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
)


// Logger provides debug, info, warn, panic, & fatal functions to log.
type Logger struct {
	internalLogger *zap.Logger
}


// LogDetails is the schema structure for logging.
type LogDetails struct {
	Message string
	ServiceId string
	CorrelationalId string
	TraceabilityId string
	ClientId string
	UserId string
	Endpoint string
	AdditionalData map[string]string
}


// KinesisHook
type KinesisHook struct {
	svc 			*kinesis.Kinesis
	Async			bool
	AcceptedLevels	[]zapcore.Level
	streamName 		string
	m				sync.Mutex
	isProd			bool
}

// InitLogger initializes a new logger.
// Connects into AWS and sets up a kinesis service.
// It returns a new logger instance and any errors upon initialization.
func InitLogger(isProd bool, loggerName, streamName, awsAccessKey, awsSecretKey, awsRegion string) (Logger, error) {
	l := Logger{}

	cred := credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, "")
	cfg := aws.NewConfig().WithRegion(awsRegion).WithCredentials(cred)

	kcHookConstructor, err := newKinesisHook(streamName, cfg, isProd)

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
		logger = logger.WithOptions(zap.Hooks(kcHook)).Named(loggerName)
		l.internalLogger = logger
	} else {
		config := zap.NewDevelopmentConfig()
		config.Encoding = "json"
		logger, _ := config.Build()
		logger = logger.WithOptions(zap.Hooks(kcHook)).Named(loggerName)
		l.internalLogger = logger
	}

	return l, nil
}


// Debug provides developer ability to send useful debug related  messages into Kinesis logging stream.
func (l *Logger) Debug(ld LogDetails) {
	j, _ := json.Marshal(ld)
	l.internalLogger.Debug(string(j))
}


// Info provides developer ability to send general info  messages into Kinesis logging stream.
func (l *Logger) Info(ld LogDetails) {
	j, _ := json.Marshal(ld)
	l.internalLogger.Info(string(j))
}


// Warn provides developer ability to send useful warning messages into Kinesis logging stream.
func (l *Logger) Warn(ld LogDetails) {
	j, _ := json.Marshal(ld)
	l.internalLogger.Warn(string(j))
}


// Fatal provides developer ability to send application fatal messages into Kinesis logging stream.
func (l *Logger) Fatal(ld LogDetails) {
	j, _ := json.Marshal(ld)
	l.internalLogger.Fatal(string(j))
}


// Error provides developer ability to send error  messages into Kinesis logging stream.
func (l *Logger) Error(ld LogDetails) {
	j, _ := json.Marshal(ld)
	l.internalLogger.Fatal(string(j))
}

// Panic provides developer ability to send panic  messages into Kinesis logging stream.
func (l *Logger) Panic(ld LogDetails) {
	j, _ := json.Marshal(ld)
	l.internalLogger.Panic(string(j))
}

// newKinesisHook creates a KinesisHook struct to to use in the zap log.
// Tries to find the existing aws Kinesis stream.
// Creates stream when doesn't exist.
// Returns a pointer with a implemented KinesisHook.
func newKinesisHook(streamName string, cfg *aws.Config, isProd bool) (*KinesisHook, error) {
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
	}

	return ks, nil
}


// getHook inserts the function to use when zap creates a log entry.
func (ch *KinesisHook) getHook() (func(zapcore.Entry) error, error) {
	kWriter := func(e zapcore.Entry) error {
		if !ch.isAcceptedLevel(e.Level) {
			return nil
		}

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

	return kWriter, nil
}


// Levels sets which levels to sent to cloudwatch
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
