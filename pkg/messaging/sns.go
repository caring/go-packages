package messaging

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

// NewSNS initializes a new AWS SNS client
func NewSNS(config *Config) (*sns.SNS, string, error) {
	c, err := mergeAndPopulateConfig(config)
	if err != nil {
		return nil, "", err
	}

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
		return nil, "", err
	}
	return client, c.TopicArn, nil
}
