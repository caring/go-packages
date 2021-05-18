package dialer

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/caring/go-packages/v2/pkg/errors"
)

// Config read from environment or connection builder.
type Config struct {
	host       string
	port       string
	caFile     string
	clientCert string
	clientKey  string
	tokenAuth  string
	basicAuth  string
	withTLS    bool
	skipVerify bool
}

// ReadConfig reads connection string from provided env variable name and returns config.
// If there are tls options the Config value can be used to generate the tls.Config
// the format is tls://hostname:port (or tcp://hostname:port if developing locally without TLS)
// Optional TLS parameters:
//   skip_verify=true        ignore server CA verification.
//   ca_file=./filename.pem  CA of service for verification.
// NOTE: The connection process is non-blocking so a timeout is not needed. To force a blocking add grpc.Blocking() to the opts
func ReadConfig(addr string) (*Config, error) {
	if addr == "" {
		return nil, errors.Errorf("connection string is empty")
	}

	var c Config
	// If there is a connection string use that.
	// format: tls://hostname:port or tcp://hostname:port
	u, err := url.Parse(addr)
	if err != nil {
		return nil, errors.Wrap(err, "connection string is malformed")
	}

	c.host = u.Hostname()
	switch u.Scheme {
	case "tcp":
		c.port, c.withTLS = "80", false
	case "tls":
		c.port, c.withTLS = "443", true
	}

	if port := u.Port(); port != "" {
		c.port = port
	}

	if c.withTLS {
		if skip := u.Query().Get("skip_verify"); skip == "true" {
			c.skipVerify = true
		} else {
			c.caFile = u.Query().Get("ca_file")
			c.clientCert = u.Query().Get("client_cert")
			c.clientKey = u.Query().Get("client_key")
		}
	}

	c.basicAuth = u.Query().Get("basic_auth")
	c.tokenAuth = u.Query().Get("token_auth")

	return &c, nil
}

func (c *Config) String() string {
	if c == nil {
		return "<nil>"
	}

	scheme := "tcp"
	if c.withTLS {
		scheme = "tls"
	}
	query := make(url.Values)
	if c.skipVerify {
		query.Set("skip_verify", "true")
	}
	if c.caFile != "" {
		query.Set("ca_file", c.caFile)
	}
	if c.clientCert != "" {
		query.Set("client_cert", c.clientCert)
	}
	if c.clientKey != "" {
		query.Set("client_key", c.clientKey)
	}
	if c.basicAuth != "" {
		query.Set("basic_auth", c.basicAuth)
	}
	if c.tokenAuth != "" {
		query.Set("token_auth", c.tokenAuth)
	}

	qry := ""
	if len(query) > 0 {
		qry = "?" + query.Encode()
	}

	return fmt.Sprintf("%s://%s:%s%s", scheme, c.host, c.port, qry)
}

// TLSConfig build config struct and load certificates if required.
func (c *Config) TLSConfig() (*tls.Config, error) {
	if c == nil {
		return nil, errors.Errorf("config is not initialized")
	}
	if !c.withTLS {
		return nil, errors.Errorf("TLS is not enabled")
	}

	tlsConfig := tls.Config{InsecureSkipVerify: c.skipVerify}

	if c.caFile != "" {
		pemServerCA, err := ioutil.ReadFile(c.caFile)
		if err != nil {
			return nil, errors.Wrap(err, "error reading CA file")
		}
		tlsConfig.RootCAs = x509.NewCertPool()
		tlsConfig.RootCAs.AppendCertsFromPEM(pemServerCA)
	}

	if c.clientCert != "" && c.clientKey != "" {
		cert, err := tls.LoadX509KeyPair(c.clientCert, c.clientKey)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read client certificate or key file")
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return &tlsConfig, nil
}

func (c *Config) ApplyToBuilder(cb *Builder) error {
	if c == nil {
		return errors.New("config is nil")
	}
	if cb == nil || cb == (*Builder)(nil) {
		return errors.New("connection builder is nil")
	}

	err := cb.SetConnInfo(c.host, c.port, c.withTLS)
	if err != nil {
		return errors.Wrap(err, "unable to set connect options")
	}
	if c.withTLS {
		tlsConfig, err := c.TLSConfig()
		if err != nil {
			return errors.Wrap(err, "unable to read tls options")
		}
		cb.WithServerTransportCredentials(tlsConfig.InsecureSkipVerify, tlsConfig.RootCAs)
		cb.WithClientTransportCredentials(tlsConfig.Certificates...)
	}
	if c.basicAuth != "" {
		cb.WithClientCredentials(authorization{authType: basicAuth, content: c.basicAuth})
	} else if c.tokenAuth != "" {
		cb.WithClientCredentials(authorization{authType: bearerAuth, content: c.tokenAuth})
	}

	return nil
}

// Authorization Types
const (
	basicAuth  authorizationType = "Basic"
	bearerAuth authorizationType = "Bearer"
)

type authorizationType string

type authorization struct {
	authType authorizationType
	content  string
}

func (s authorization) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	enc := base64.StdEncoding.EncodeToString([]byte(s.content))
	return map[string]string{
		"authorization": fmt.Sprintf("%s %s", s.authType, enc),
	}, nil
}

func (authorization) RequireTransportSecurity() bool {
	return true
}
