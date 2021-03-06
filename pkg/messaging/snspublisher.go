package messaging

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/caring/go-packages/v2/pkg/logging"
)

const (
	Subject = "Subject"
)

func Publish(client *sns.SNS, logger *logging.Logger, subject, topicArn, json string) (string, error) {
	var err error
	if client == nil || len(topicArn) == 0 {
		client, err = NewSNS(&Config{
			Logger: logger,
		})
		if err != nil {
			return "", err
		}
	}

	messageAttributeValue := sns.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(subject),
	}
	messageAttributes := map[string]*sns.MessageAttributeValue{Subject: &messageAttributeValue}

	input := sns.PublishInput{
		Message:           aws.String(json),
		MessageAttributes: messageAttributes,
		Subject:           aws.String(subject),
		TopicArn:          aws.String(topicArn),
	}

	result, err := client.Publish(&input)
	if err != nil {
		return "", err
	}

	return *result.MessageId, err
}
