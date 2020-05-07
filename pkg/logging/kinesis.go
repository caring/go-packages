package logging

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"go.uber.org/zap/zapcore"
)

// AllLevels Supported log levels
var AllLevels = []zapcore.Level{
	zapcore.DebugLevel,
	zapcore.InfoLevel,
	zapcore.WarnLevel,
	zapcore.ErrorLevel,
	zapcore.FatalLevel,
	zapcore.PanicLevel,
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
