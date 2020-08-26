package messaging

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// NewSQS initializes a new AWS SQS client
func NewSQS(config *Config) (*sqs.SQS, error) {
	c, err := mergeAndPopulateConfig(config)
	if err != nil {
		return nil, err
	}
	l := c.Logger

	credVal := credentials.Value{
		AccessKeyID:     c.AccessKeyID,
		SecretAccessKey: c.SecretAccessKey,
	}
	cred := credentials.NewStaticCredentialsFromCreds(credVal)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: cred,
			Region:      aws.String(c.AWSRegion),
		},
		SharedConfigState: session.SharedConfigEnable,
	}))

	client := sqs.New(sess)
	if client == nil {
		l.Fatal("Failed to establish connection to SQS")
		return nil, err
	}
	return client, nil
}
