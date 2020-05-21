package logging

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"go.uber.org/zap/zapcore"
)

// KinesisHook providers getters and generators for hooks to connect zap to kinesis
type kinesisHook struct {
	svc          *kinesis.Kinesis
	partitionKey string
	streamName   string
}

// newKinesisHook creates a KinesisHook, returns an error if the configured Kinesis stream does not exist
func newKinesisHook(streamName string, partitionKey string) (*kinesisHook, error) {
	s := session.New()
	kc := kinesis.New(s)

	_, err := kc.DescribeStream(&kinesis.DescribeStreamInput{StreamName: aws.String(streamName)})

	// Create stream if doesn't exist
	if err != nil {
		return nil, err
	}

	ks := &kinesisHook{
		streamName: streamName,
		svc:        kc,
	}

	return ks, nil
}

// TODO Zap docs state not to use hooks for more complex solutions like another logging output,
// kinesis errors get swallowed here. We need to implement kinesis logging a zap.Core, where we have more
// control on error destination, log level and good concurrency models

// getHook provides the writer hook to zap that will call to write log entries to kinesis
func (ch *kinesisHook) getHook() (func(zapcore.Entry) error, error) {
	kWriter := func(e zapcore.Entry) error {
		writer := func() error {
			_, err := ch.svc.PutRecord(&kinesis.PutRecordInput{
				Data:         []byte(e.Message),
				StreamName:   aws.String(ch.streamName),
				PartitionKey: &ch.partitionKey,
			})

			if err != nil {
				return err
			}

			return err
		}

		go writer()
		return nil
	}

	return kWriter, nil
}
