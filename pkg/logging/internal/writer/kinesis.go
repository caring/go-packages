package writer

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/google/uuid"
)

type kinesisWriter struct {
	*kinesis.Kinesis
	streamName string
}

func newKinesisWriter(streamName string) (io.Writer, error) {
	ses, err := session.NewSession(&aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	ks := kinesis.New(ses)

	_, err = ks.DescribeStream(&kinesis.DescribeStreamInput{StreamName: aws.String(streamName)})

	if err != nil {
		return nil, err
	}

	return &kinesisWriter{ks, streamName}, nil

}

func (k *kinesisWriter) Write(p []byte) (n int, err error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return 0, err
	}
	key := id.String()

	_, err = k.PutRecord(&kinesis.PutRecordInput{
		Data:         p,
		StreamName:   aws.String(k.streamName),
		PartitionKey: &key,
	})
	if err != nil {
		return 0, err
	}

	return len(p), nil
}
