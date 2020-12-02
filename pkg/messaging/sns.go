package messaging

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

// NewSNS initializes a new AWS SNS client
func NewSNS(config *Config) (*sns.SNS, error) {
	c, err := mergeAndPopulateConfig(config)
	if err != nil {
		return nil, err
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String(c.AWSRegion),
		},
		SharedConfigState: session.SharedConfigEnable,
	}))
	creds := stscreds.NewCredentials(sess, c.RoleArn)
	client := sns.New(sess, &aws.Config{Credentials: creds})
	if client == nil {
		return nil, err
	}
	return client, nil
}
