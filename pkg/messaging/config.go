package messaging

import (
	"errors"
	"github.com/caring/go-packages/pkg/logging"
	"os"
)

// Config contains initialization config for NewTracer
type Config struct {
	// AWS Region
	AWSRegion string
	// AccessKeyID
	AccessKeyID string
	// SecretAccessKey
	SecretAccessKey string
	// TopicArn for EmailTopic
	TopicArn string
	// The instance of our own logger to use for logging traces
	Logger *logging.Logger
}

func newDefaultConfig() *Config {
	return &Config{
		AWSRegion:       "",
		AccessKeyID:     "",
		SecretAccessKey: "",
		TopicArn:        "",
		Logger:          nil,
	}
}

// mergeAndPopulateConfig starts with a default config, and populates
// it with config from the environment. Config from the environment can
// be overridden with any config input as arguments. Only non 0 values will
// overwrite the defaults
func mergeAndPopulateConfig(c *Config) (*Config, error) {
	final := newDefaultConfig()

	if c.Logger == nil {
		return nil, errors.New("No logger input")
	}
	final.Logger = c.Logger

	if c.AWSRegion != "" {
		final.AWSRegion = c.AWSRegion
	} else if s := os.Getenv("AWS_REGION"); s != "" {
		final.AWSRegion = s
	}

	if c.AccessKeyID != "" {
		final.AccessKeyID = c.AccessKeyID
	} else if s := os.Getenv("ACCESS_KEY_ID"); s != "" {
		final.AccessKeyID = s
	}

	if c.SecretAccessKey != "" {
		final.SecretAccessKey = c.SecretAccessKey
	} else if s := os.Getenv("SECRET_ACCESS_KEY"); s != "" {
		final.SecretAccessKey = s
	}

	if c.TopicArn != "" {
		final.TopicArn = c.TopicArn
	} else if s := os.Getenv("EMAIL_TOPIC_ARN"); s != "" {
		final.TopicArn = s
	}

	return final, nil
}
