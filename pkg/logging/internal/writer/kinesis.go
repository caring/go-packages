package writer

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
)

type kinesisWriter struct {
	*firehose.Firehose
	streamName string
}

// NewKinesisWriter creates an io.Writer that will write to the given kinesis stream name.NewKinesisWriter
// All other AWS configuration is picked up from the runtime hardware via environnement variables. See AWS docs
func NewKinesisWriter(streamName string) (io.Writer, error) {
	ses, err := session.NewSession(&aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	h := firehose.New(ses)

	_, err = h.DescribeDeliveryStream(&firehose.DescribeDeliveryStreamInput{DeliveryStreamName: aws.String(streamName)})

	if err != nil {
		return nil, err
	}

	return &kinesisWriter{h, streamName}, nil

}

// Write writes one byte slice as one kinesis record to a random shard,
// and blocks until the response is returned
func (k *kinesisWriter) Write(p []byte) (n int, err error) {
	_, err = k.PutRecord(&firehose.PutRecordInput{
		Record: &firehose.Record{
			Data: p,
		},
		DeliveryStreamName: aws.String(k.streamName),
	})
	if err != nil {
		return 0, err
	}

	return len(p), nil
}
