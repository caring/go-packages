package messaging

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
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

	var client *sns.SNS
	if c.RoleArn != "" {
		sess, err := session.NewSession(&aws.Config{Region: aws.String(c.AWSRegion)})
		if err != nil {
			return nil, err
		}
		creds := stscreds.NewCredentials(sess, c.RoleArn)
		client = sns.New(sess, &aws.Config{Credentials: creds})
	} else if c.AccessKeyID != "" && c.SecretAccessKey != "" {
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
		client = sns.New(sess)
	} else {
		sess, err := session.NewSession(&aws.Config{Region: aws.String(c.AWSRegion)})
		if err != nil {
			return nil, err
		}
		client = sns.New(sess)
	}
	if client == nil {
		return nil, err
	}
	return client, nil
}
