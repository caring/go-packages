package messaging

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"strings"
)

const (
	EmailTopic = "EmailTopic"
)

// NewSNS initializes a new AWS SNS client
func NewSNS(config *Config) (*sns.SNS, string, error) {
	c, err := mergeAndPopulateConfig(config)
	if err != nil {
		return nil, "", err
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

	client := sns.New(sess)
	if client == nil {
		l.Fatal("Failed to establish connection to SNS")
		return nil, "", err
	}

	result, err := client.ListTopics(nil)
	if err != nil {
		l.Fatal("Failed to list SNS topics:" + err.Error())
		return nil, "", err
	}

	topicArn := ""
	for _, t := range result.Topics {
		if strings.Contains(*t.TopicArn, EmailTopic) {
			topicArn = *t.TopicArn
			break
		}
	}
	return client, topicArn, nil
}
