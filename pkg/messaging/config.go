package messaging

import (
	"errors"
	"github.com/caring/go-packages/v2/pkg/logging"
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
	// RoleArn
	RoleArn string
	// The instance of our own logger to use for logging traces
	Logger *logging.Logger
}

func newDefaultConfig() *Config {
	return &Config{
		AWSRegion:       "",
		AccessKeyID:     "",
		SecretAccessKey: "",
		RoleArn:         "",
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

	if s := os.Getenv("AWS_REGION"); s != "" {
		final.AWSRegion = s
	} else if c.AWSRegion != "" {
		final.AWSRegion = c.AWSRegion
	} else {
		return nil, errors.New("Missing environment variable AWS_REGION")
	}

	if s := os.Getenv("AWS_ROLE_ARN"); s != "" {
		final.RoleArn = s
	} else if c.RoleArn != "" {
		final.RoleArn = c.RoleArn
	}

	if s := os.Getenv("AWS_ACCESS_KEY_ID"); s != "" {
		final.AccessKeyID = s
	} else if c.AccessKeyID != "" {
		final.AccessKeyID = c.AccessKeyID
	}

	if s := os.Getenv("AWS_SECRET_ACCESS_KEY"); s != "" {
		final.SecretAccessKey = s
	} else if c.SecretAccessKey != "" {
		final.SecretAccessKey = c.SecretAccessKey
	}

	return final, nil
}
