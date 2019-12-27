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
	serviceId string
	writeToKinesis bool
}


// LogDetails is the schema structure for logging.
type LogDetails struct {
	Message string
	CorrelationalId string
	TraceabilityId string
	ClientId string
	UserId string
	Endpoint string
	AdditionalData map[string]interface{}
	IsReportable *bool
	ParentDetails *LogDetails
	Logger *Logger
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


// InitLogger initializes a new logger.
// Connects into AWS and sets up a kinesis service.
// It returns a new logger instance and any errors upon initialization.
func InitLogger(isProd bool, loggerName, streamName, serviceID, awsAccessKey, awsSecretKey, awsRegion string, isAsync, writeToKinesis bool) (Logger, error) {
	l := Logger{}
	l.serviceId = serviceID

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


// Debug provides developer ability to send useful debug related  messages into Kinesis logging stream.
func (d *LogDetails) Debug(message string) error {
	if message != "" {
		d.Message = message
	}

	c, err := d.getLogContent()
	d.Logger.GetInternalLogger().Debug(c)

	return err
}


// Info provides developer ability to send general info  messages into Kinesis logging stream.
func (d *LogDetails) Info(message string) error {
	if message != "" {
		d.Message = message
	}

	c, err := d.getLogContent()
	d.Logger.GetInternalLogger().Info(c)

	return err
}


// Warn provides developer ability to send useful warning messages into Kinesis logging stream.
func (d *LogDetails) Warn(message string) error {
	if message != "" {
		d.Message = message
	}

	c, err := d.getLogContent()
	d.Logger.GetInternalLogger().Warn(c)

	return err
}


// Fatal provides developer ability to send application fatal messages into Kinesis logging stream.
func (d *LogDetails) Fatal(message string) error {
	if message != "" {
		d.Message = message
	}

	c, err := d.getLogContent()
	d.Logger.GetInternalLogger().Fatal(c)

	return err
}


// Error provides developer ability to send error  messages into Kinesis logging stream.
func (d *LogDetails) Error(message string) error {
	if message != "" {
		d.Message = message
	}

	c, err := d.getLogContent()
	d.Logger.GetInternalLogger().Error(c)

	return err
}


// Panic provides developer ability to send panic  messages into Kinesis logging stream.
func (d *LogDetails) Panic(message string) error {
	if message != "" {
		d.Message = message
	}

	c, err := d.getLogContent()
	d.Logger.GetInternalLogger().Panic(c)

	return err
}


func (d *LogDetails) updateEmptyWithParent() {
	if d.IsReportable == nil {
		d.IsReportable = d.ParentDetails.IsReportable
	}

	if d.ClientId == "" {
		d.ClientId = d.ParentDetails.ClientId
	}

	if d.CorrelationalId == "" {
		d.CorrelationalId = d.ParentDetails.CorrelationalId
	}

	if d.Endpoint == "" {
		d.Endpoint = d.ParentDetails.Endpoint
	}

	if d.UserId == "" {
		d.UserId = d.ParentDetails.UserId
	}

	if d.Message == "" {
		d.Message = d.ParentDetails.Message
	}

	if d.TraceabilityId == "" {
		d.TraceabilityId = d.ParentDetails.TraceabilityId
	}

	if d.AdditionalData == nil {
		d.AdditionalData = d.ParentDetails.AdditionalData
	}

	if d.hasEmpty() && d.ParentDetails.ParentDetails != nil {
		d.ParentDetails = d.ParentDetails.ParentDetails
		d.updateEmptyWithParent()
	}
}


func (d *LogDetails) hasEmpty() bool {
	hasEmpty := d.IsReportable == nil || d.Message == "" || d.ClientId == "" || d.CorrelationalId == "" || d.Endpoint == "" || d.TraceabilityId == "" || d.UserId == "" || d.AdditionalData == nil

	return hasEmpty
}


// getLogContents aggregates the LogDetails and Logger into a combined map.
// It returns a json string to insert into an actual log.
func (d *LogDetails) getLogContent() (string, error) {
	d.updateEmptyWithParent()

	m := map[string]interface{}{
		"message":         d.Message,
		"additionalData":  d.AdditionalData,
		"userID":          d.UserId,
		"traceabilityID":  d.TraceabilityId,
		"endpoint":        d.Endpoint,
		"correlationalID": d.CorrelationalId,
		"clientID":        d.ClientId,
		"serviceID":       d.Logger.serviceId,
		"isReportable":    d.IsReportable,
	}

	jc, err := json.Marshal(m)

	if err != nil {
		return "", err
	}

	return string(jc), nil
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
