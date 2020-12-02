package messaging

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// NewSQS initializes a new AWS SQS client
func NewSQS(config *Config) (*sqs.SQS, error) {
	c, err := mergeAndPopulateConfig(config)
	if err != nil {
		return nil, err
	}

	var client *sqs.SQS
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	awscfg := &aws.Config{Region: aws.String(c.AWSRegion)}
	if c.RoleArn != "" {
		creds := stscreds.NewCredentials(sess, c.RoleArn)
		awscfg.Credentials = creds
	} else if c.AccessKeyID != "" && c.SecretAccessKey != "" {
		credVal := credentials.Value{
			AccessKeyID:     c.AccessKeyID,
			SecretAccessKey: c.SecretAccessKey,
		}
		cred := credentials.NewStaticCredentialsFromCreds(credVal)
		awscfg.Credentials = cred
		sess = session.Must(session.NewSessionWithOptions(session.Options{
			Config: aws.Config{
				Credentials: cred,
				Region:      aws.String(c.AWSRegion),
			},
			SharedConfigState: session.SharedConfigEnable,
		}))
	}
	client = sqs.New(sess, awscfg)
	if client == nil {
		return nil, err
	}
	return client, nil
}
