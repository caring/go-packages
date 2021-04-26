package messaging

import (
	"github.com/aws/aws-sdk-go/aws"
	_ "github.com/aws/aws-sdk-go/aws/credentials"
	_ "github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"strings"
)

// NewSNS initializes a new AWS SNS client
func NewSNS(config *Config) (*sns.SNS, error) {
	c, err := mergeAndPopulateConfig(config)
	if err != nil {
		return nil, err
	}

	awscfg := &aws.Config{
		Region:                        aws.String(c.AWSRegion),
		CredentialsChainVerboseErrors: aws.Bool(true),
	}

	sess, err := session.NewSession(awscfg)
	if err != nil {
		return nil, err
	}

	var client *sns.SNS
	client = sns.New(sess)
	if client == nil {
		return nil, err
	}
	return client, nil
}

// TopicArn returns topicArn for a specific topic display name
func TopicArn(client *sns.SNS, topic string) (string, error) {
	topicArn := ""
	if len(topic) > 0 {
		topics, err := client.ListTopics(nil)
		if err != nil {
			return "", err
		}
		for _, t := range topics.Topics {
			if strings.Contains(*t.TopicArn, topic) {
				topicArn = *t.TopicArn
				break
			}
		}
	}
	return topicArn, nil
}
