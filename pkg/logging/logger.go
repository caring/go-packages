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
	writeToKinesis bool
}


// LogDetails is the schema structure for logging.
type LogDetails struct {
	serviceID string
	correlationalId string
	traceabilityId string
	clientId string
	userId string
	endpoint string
	additionalData map[string]interface{}
	isReportable *bool
	parentDetails *LogDetails
	logger *Logger
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
		isReportable:    nil,
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


// GetInternalLogger returns the zap internal logger pointer
func (l *Logger) GetInternalLogger() *zap.Logger {
	return l.internalLogger
}


func (d *LogDetails) NewChild(serviceID, correlationalID, traceabilityID, clientID, userID, endpoint string, isReportable *bool, additionalData map[string]interface{}) *LogDetails {
	ld := d
	ld.parentDetails = d

	if serviceID != "" {
		ld.serviceID = serviceID
	}

	if correlationalID != "" {
		ld.correlationalId = correlationalID
	}

	if traceabilityID != "" {
		ld.traceabilityId = traceabilityID
	}

	if clientID != "" {
		ld.clientId = clientID
	}

	if userID != "" {
		ld.userId = userID
	}

	if isReportable != nil {
		ld.isReportable = isReportable
	}

	if additionalData != nil {
		ld.additionalData = additionalData
	}

	return ld
}

func (d *LogDetails) SetServiceID(serviceID string) *LogDetails {
	d.serviceID = serviceID

	return d
}


func (d *LogDetails) SetCorrelationalID(correlationalID string) *LogDetails {
	d.correlationalId = correlationalID

	return d
}

func (d *LogDetails) SetClientID(clientID string) *LogDetails {
	d.clientId = clientID

	return d
}

func (d *LogDetails) SetTraceabilityID(traceabilityID string) *LogDetails {
	d.traceabilityId = traceabilityID

	return d
}

func (d *LogDetails) SetUserID(userID string) *LogDetails {
	d.userId = userID

	return d
}

func (d *LogDetails) SetIsReportable(isReportable *bool) *LogDetails {
	d.isReportable = isReportable

	return d
}


func (d *LogDetails) SetAdditionalData(additionalData map[string]interface{}) *LogDetails {
	d.additionalData = additionalData

	return d
}


// Debug provides developer ability to send useful debug related  messages into Kinesis logging stream.
func (d *LogDetails) Debug(message string, additionalData map[string]interface{}) error {
	c, err := d.getLogContent(message, additionalData)
	d.logger.GetInternalLogger().Debug(c)

	return err
}


// Info provides developer ability to send general info  messages into Kinesis logging stream.
func (d *LogDetails) Info(message string, additionalData map[string]interface{}) error {
	c, err := d.getLogContent(message, additionalData)
	d.logger.GetInternalLogger().Info(c)

	return err
}


// Warn provides developer ability to send useful warning messages into Kinesis logging stream.
func (d *LogDetails) Warn(message string, additionalData map[string]interface{}) error {
	c, err := d.getLogContent(message, additionalData)
	d.logger.GetInternalLogger().Warn(c)

	return err
}


// Fatal provides developer ability to send application fatal messages into Kinesis logging stream.
func (d *LogDetails) Fatal(message string, additionalData map[string]interface{}) error {
	c, err := d.getLogContent(message, additionalData)
	d.logger.GetInternalLogger().Fatal(c)

	return err
}


// Error provides developer ability to send error  messages into Kinesis logging stream.
func (d *LogDetails) Error(message string, additionalData map[string]interface{}) error {
	c, err := d.getLogContent(message, additionalData)
	d.logger.GetInternalLogger().Error(c)

	return err
}


// Panic provides developer ability to send panic  messages into Kinesis logging stream.
func (d *LogDetails) Panic(message string, additionalData map[string]interface{}) error {
	c, err := d.getLogContent(message, additionalData)
	d.logger.GetInternalLogger().Panic(c)

	return err
}


func (d *LogDetails) hasEmpty() bool {
	hasEmpty := d.isReportable == nil || d.clientId == "" || d.correlationalId == "" || d.endpoint == "" || d.traceabilityId == "" || d.userId == "" || d.additionalData == nil

	return hasEmpty
}


// getLogContents aggregates the LogDetails and Logger into a combined map.
// It returns a json string to insert into an actual log.
func (d *LogDetails) getLogContent(message string, additionalData map[string]interface{}) (string, error) {
	if d.additionalData == nil {
		d.additionalData = additionalData
	} else if additionalData != nil {
		for k, v := range additionalData {
			d.additionalData[k] = v
		}
	}

	m := map[string]interface{}{
		"message":         message,
		"additionalData":  d.additionalData,
		"userID":          d.userId,
		"traceabilityID":  d.traceabilityId,
		"endpoint":        d.endpoint,
		"correlationalID": d.correlationalId,
		"clientID":        d.clientId,
		"serviceID":       d.serviceID,
		"isReportable":    d.isReportable,
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
