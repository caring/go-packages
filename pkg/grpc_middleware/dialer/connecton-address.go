package dialer

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/caring/go-packages/v2/pkg/errors"
)

// ConnectionAddress read from environment or connection builder.
type ConnectionAddress struct {
	host       string
	port       string
	caFile     string
	clientCert string
	clientKey  string
	tokenAuth  string
	basicAuth  string
	disableTLS bool
	skipVerify bool
}

// ReadConnectionAddress reads connection string from provided env variable name and returns config.
// If there are tls options the ConnectionAddress value can be used to generate the tls.ConnectionAddress
// the format is tls://hostname:port (or tcp://hostname:port if developing locally without TLS)
// Optional TLS parameters:
//   skip_verify=true             ignore server CA verification.
//   ca_file=./filename.pem       CA of service for verification.
//   client_cert=./filename.pem   Client certificate to use for authentication.
//   client_key=./filename.pem    Client private key to use for authentication.
//   basic_auth=user:pass         Username and Password for basic authentication.
//   token_auth=token             Bearer token for token authentication.
// NOTE: The connection process is non-blocking so a timeout is not needed. To force a blocking add grpc.Blocking() to the opts
func ReadConnectionAddress(addr string) (*ConnectionAddress, error) {
	var c ConnectionAddress

	if addr == "" {
		if c.host == "" || c.port == "" {
			return nil, errors.Errorf("connection string is empty")
		}

		return &c, nil
	}

	// Remove dns: if present
	addr = strings.TrimPrefix(addr, "dns:")

	// add tls scheme if no scheme is present.
	if !strings.Contains(addr, "://") {
		addr = "tls://" + addr
	}

	// If there is a connection string use that.
	// format: tls://hostname:port or tcp://hostname:port
	u, err := url.Parse(addr)
	if err != nil {
		return nil, errors.Wrap(err, "connection string is malformed")
	}

	c.host = u.Hostname()
	if c.host == "" {
		return nil, errors.Wrap(err, "connection string has no hostname")
	}

	switch u.Scheme {
	case "tcp", "http":
		c.port, c.disableTLS = "80", true
	case "tls", "https":
		c.port, c.disableTLS = "443", false
	}

	if port := u.Port(); port != "" {
		c.port = port
	}

	if !c.disableTLS {
		if skip := u.Query().Get("skip_verify"); skip == "true" {
			c.skipVerify = true
		} else {
			c.caFile = u.Query().Get("ca_file")
		}
	}

	return &c, nil
}

func (c *ConnectionAddress) String() string {
	if c == nil {
		return "<nil>"
	}

	scheme := "tls"
	if c.disableTLS {
		scheme = "tcp"
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

func (c *ConnectionAddress) loadTLS(fs fs.FS) (*tls.Config, error) {
	if c == nil {
		return nil, errors.Errorf("config is not initialized")
	}
	if c.disableTLS {
		return nil, errors.Errorf("TLS is not enabled")
	}

	tlsConnectionAddress := tls.Config{InsecureSkipVerify: c.skipVerify}

	if c.caFile != "" {
		pemServerCA, err := ioutil.ReadFile(c.caFile)
		if err != nil {
			return nil, errors.Wrap(err, "error reading CA file")
		}
		tlsConnectionAddress.RootCAs = x509.NewCertPool()
		tlsConnectionAddress.RootCAs.AppendCertsFromPEM(pemServerCA)
	}

	if c.clientCert != "" && c.clientKey != "" {
		cert, err := tls.LoadX509KeyPair(c.clientCert, c.clientKey)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read client certificate or key file")
		}

		tlsConnectionAddress.Certificates = []tls.Certificate{cert}
	}

	return &tlsConnectionAddress, nil
}

func (c *ConnectionAddress) ApplyToBuilder(cb *Builder) error {
	if c == nil {
		return errors.New("config is nil")
	}
	if cb == nil || cb == (*Builder)(nil) {
		return errors.New("connection builder is nil")
	}

	err := cb.SetConnInfo(c.host, c.port, !c.disableTLS)
	if err != nil {
		return errors.Wrap(err, "unable to set connect options")
	}
	if !c.disableTLS {
		fs := cb.GetFS()

		tlsConnectionAddress, err := c.loadTLS(fs)
		if err != nil {
			return errors.Wrap(err, "unable to read tls options")
		}
		cb.WithServerTransportCredentials(tlsConnectionAddress.InsecureSkipVerify, tlsConnectionAddress.RootCAs)
		cb.WithClientTransportCredentials(tlsConnectionAddress.Certificates...)
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
