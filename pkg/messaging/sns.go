package messaging

import (
	"github.com/aws/aws-sdk-go/aws"
	_ "github.com/aws/aws-sdk-go/aws/credentials"
	_ "github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"strings"
)

// NewSNS initializes a new AWS SNS client and returns topicArn - default for EmailTopic topic
func NewSNS(config *Config, params ...string) (*sns.SNS, string, error) {
	c, err := mergeAndPopulateConfig(config)
	if err != nil {
		return nil, "", err
	}

	awscfg := &aws.Config{
		Region:                        aws.String(c.AWSRegion),
		CredentialsChainVerboseErrors: aws.Bool(true),
	}

	sess, err := session.NewSession(awscfg)
	if err != nil {
		return nil, "", err
	}

	var client *sns.SNS
	client = sns.New(sess)
	if client == nil {
		return nil, "", err
	}

	topic := "EmailTopic"
	if len(params) > 0 {
		topic = params[0]
	}
	topics, err := client.ListTopics(nil)
	if err != nil {
		return nil, "", err
	}

	topicArn := ""
	for _, t := range topics.Topics {
		if strings.Contains(*t.TopicArn, topic) {
			topicArn = *t.TopicArn
			break
		}
	}
	return client, topicArn, nil
}
