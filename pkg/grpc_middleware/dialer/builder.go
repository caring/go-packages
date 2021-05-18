package dialer

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/caring/go-packages/v2/pkg/errors"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

const (
	// MinConnectTimeout before attempting reconnect.
	MinConnectTimeout = 20 * time.Second
)

type Builder struct {
	options         []grpc.DialOption
	enabledBlocking bool
	connectParams   grpc.ConnectParams
	keepAliveParams keepalive.ClientParameters
	credentials     credentials.PerRPCCredentials
	tlsConfig       *tls.Config
	uinterceptors   []grpc.UnaryClientInterceptor
	sinterceptors   []grpc.StreamClientInterceptor
	dns             *string
	port            *string
	withTLS         *bool
}

// WithOptions allows possing in multiple grpc dial options
// DialOption configures how we set up the connection.
func (b *Builder) WithOptions(opts ...grpc.DialOption) {
	b.options = opts
}

// AppendOptions appends the given options to the current set of options.
func (b *Builder) AppendOptions(opts ...grpc.DialOption) {
	b.options = append(b.options, opts...)
}

// GetOptions returns the current set of options
func (b *Builder) GetOptions() []grpc.DialOption {
	return b.options
}

// WithBlock set to true applies a DialOption which makes caller of Dial blocks until the
// underlying connection is up. Without this, Dial returns immediately and
// connecting the server happens in background.
func (b *Builder) WithBlock(blocking bool) {
	b.enabledBlocking = blocking
}

// GetBlock returns current value
func (b *Builder) GetBlock() bool {
	return b.enabledBlocking
}

// DefaultBackoff retries its connection to service if it fails to connect.
// DefaultConfig is a backoff configuration with the default values specfied
// at https://github.com/grpc/grpc/blob/master/doc/connection-backoff.md.
func (b *Builder) WithDefaultBackoff() {
	b.WithConnectParams(grpc.ConnectParams{
		MinConnectTimeout: MinConnectTimeout,
		Backoff:           backoff.DefaultConfig,
	})
}

// WithConnectParams sets connection parameters such as backoff and timeout.
func (b *Builder) WithConnectParams(params grpc.ConnectParams) {
	b.connectParams = params
}

// GetConnectParams returns the current configured params
func (b *Builder) GetConnectParams() grpc.ConnectParams {
	return b.connectParams
}

// WithKeepAliveParams set the keep alive params
// ClientParameters is used to set keepalive parameters on the client-side.
// These configure how the client will actively probe to notice when a connection
// is broken and send pings so intermediaries will be aware of the liveness of
// the connection. Make sure these parameters are set in coordination with the
// keepalive policy on the server, as incompatible settings can result in closing
// of connection.
func (b *Builder) WithKeepAliveParams(params keepalive.ClientParameters) {
	b.keepAliveParams = params
}

// GetKeepAliveParams returns the current keep alive params
func (b *Builder) GetKeepAliveParams() keepalive.ClientParameters {
	return b.keepAliveParams
}

// WithUnaryInterceptors set a list of interceptors to the Grpc client for unary
// connection. We leverage grpc_middleware package ChainUnaryClient to creates a
// single interceptor out of a chain of many interceptors to override the grpc
// default behavior of only allowing one.
func (b *Builder) WithUnaryInterceptors(interceptors ...grpc.UnaryClientInterceptor) {
	b.uinterceptors = interceptors
}

// AppendUnaryInterceptors ...
func (b *Builder) AppendUnaryInterceptors(interceptors ...grpc.UnaryClientInterceptor) {
	b.uinterceptors = append(b.uinterceptors, interceptors...)
}

// GetUnaryInterceptors returns the UnaryClientInterceptors slice
func (b *Builder) GetUnaryInterceptors() []grpc.UnaryClientInterceptor {
	return b.uinterceptors
}

// WithStreamInterceptors set a list of interceptors to the Grpc client for unary
// connection. We leverage grpc_middleware package ChainStreamClient to creates a
// single interceptor out of a chain of many interceptors to override the grpc
// default behavior of only allowing one.
func (b *Builder) WithStreamInterceptors(interceptors ...grpc.StreamClientInterceptor) {
	b.sinterceptors = interceptors
}

// AppendStreamInterceptors ...
func (b *Builder) AppendStreamInterceptors(interceptors ...grpc.StreamClientInterceptor) {
	b.sinterceptors = append(b.sinterceptors, interceptors...)
}

// GetStreamInterceptors returns the StreamClientInterceptors slice
func (b *Builder) GetStreamInterceptors() []grpc.StreamClientInterceptor {
	return b.sinterceptors
}

// WithServerTransportCredentials builds transport credentials for the service a gRPC client connects with
func (b *Builder) WithServerTransportCredentials(insecureSkipVerify bool, certPool *x509.CertPool) {
	if b.tlsConfig == nil {
		b.tlsConfig = &tls.Config{}
	}
	if insecureSkipVerify {
		b.tlsConfig.InsecureSkipVerify = true
	}
	if certPool != nil {
		b.tlsConfig.RootCAs = certPool
	}
}

// GetServerTransportCredentials ...
func (b *Builder) GetServerTransportCredentials() (bool, *x509.CertPool) {
	return b.tlsConfig.InsecureSkipVerify, b.tlsConfig.RootCAs
}

// WithClientTransportCredentials builds transport credentials for a gRPC client uses to connect to service
func (b *Builder) WithClientTransportCredentials(certs ...tls.Certificate) {
	b.tlsConfig.Certificates = certs
}

// GetClientTransportCredentials ...
func (b *Builder) GetClientTransportCredentials() []tls.Certificate {
	return b.tlsConfig.Certificates
}

// WithClientCredentials builds transport credentials for a gRPC client
func (b *Builder) WithClientCredentials(c credentials.PerRPCCredentials) {
	b.credentials = c
}

// GetClientCredentials ...
func (b *Builder) GetClientCredentials() credentials.PerRPCCredentials {
	return b.credentials
}

// SetConnInfo allows passing in the dns and port for the connection, providing
// flexibility if a consumer wants to set a default and or override
func (b *Builder) SetConnInfo(dns, port string, tls bool) error {
	u, err := url.Parse(dns)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("invalid dns: %s", dns))
	}
	if u.Scheme != "" {
		// This will cause the connection to fail silently with a cryptic "context.Deadline exceeded" so we validate for it here.
		return errors.New(fmt.Sprintf("grpc connection dns must not contain a scheme/protocol: %s", dns))
	}
	b.dns = &dns

	i, err := strconv.ParseUint(port, 10, 32)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("invalid port number: %s", port))
	}
	if i > 65535 {
		return errors.Errorf("invalid port number: %s", port)
	}
	b.port = &port
	b.withTLS = &tls

	return nil
}

// GetConnInfo returns the dns and port that were set, and errors if either
// were null, allowing verfication prior to attempting the connection
func (b *Builder) GetConnInfo() (dns string, port string, useTLS bool, err error) {
	if b.dns == nil || b.port == nil || b.withTLS == nil {
		return "", "", false, errors.New("Connection info not fully set")
	}
	return *b.dns, *b.port, *b.withTLS, nil
}

// SetConnAddr sets connection information as well as configures some common tls and authentication options
// based on a provided connection string.
// the format is tls://hostname:port (or tcp://hostname:port if developing locally without TLS)
// Optional TLS parameters:
//   skip_verify=true        ignore server CA verification.
//   ca_file=./filename.pem  CA of service for verification.
func (b *Builder) SetConnAddr(addr string) error {
	cfg, err := ReadConfig(addr)
	if err != nil {
		return err
	}
	cfg.ApplyToBuilder(b)
	if err != nil {
		return err
	}
	return nil
}

// Dial returns the client connection to the server.
// context is ignored unless builder is set to block using WithBlock(true)
func (b *Builder) Dial(ctx context.Context) (*grpc.ClientConn, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	dns, port, withTLS, err := b.GetConnInfo()
	if err != nil {
		return nil, fmt.Errorf("target connection parameter missing: dns and/or port not set")
	}

	addr := dns + ":" + port
	log.Printf("Target to connect = %s, tls = %t", addr, withTLS)

	options := b.joinOptions(withTLS)

	cc, err := grpc.DialContext(ctx, addr, options...)

	if err != nil {
		return nil, fmt.Errorf("unable to connect to client. error = %+v", err)
	}
	return cc, nil
}

// DialAddr returns the client connection to the server
// context is ignored unless builder is set to block using WithBlock(true)
func (b *Builder) DialAddr(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	if err := b.SetConnAddr(addr); err != nil {
		return nil, err
	}

	return b.Dial(ctx)
}

// Clone the builder and return copy
func (b *Builder) Clone() *Builder {
	c := &Builder{
		options:         make([]grpc.DialOption, len(b.options)),
		enabledBlocking: b.enabledBlocking,
		connectParams:   b.connectParams,
		credentials:     b.credentials,
		keepAliveParams: b.keepAliveParams,
		uinterceptors:   make([]grpc.UnaryClientInterceptor, len(b.uinterceptors)),
		sinterceptors:   make([]grpc.StreamClientInterceptor, len(b.sinterceptors)),
	}
	copy(c.options, b.options)
	copy(c.uinterceptors, b.uinterceptors)
	copy(c.sinterceptors, b.sinterceptors)

	if b.dns != nil {
		v := *b.dns
		c.dns = &v
	}
	if b.port != nil {
		v := *b.port
		c.port = &v
	}
	if b.withTLS != nil {
		v := *b.withTLS
		c.withTLS = &v
	}
	if b.tlsConfig != nil {
		c.tlsConfig = &tls.Config{
			Certificates: make([]tls.Certificate, len(b.tlsConfig.Certificates)),
			RootCAs:      b.tlsConfig.RootCAs,
		}
		copy(c.tlsConfig.Certificates, b.tlsConfig.Certificates)
	}
	return c
}

func (b *Builder) joinOptions(withTLS bool) []grpc.DialOption {
	var options []grpc.DialOption
	if b.enabledBlocking {
		options = append(options, grpc.WithBlock())
	}
	options = append(options, grpc.WithKeepaliveParams(b.keepAliveParams))
	options = append(options, grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(b.uinterceptors...)))
	options = append(options, grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(b.sinterceptors...)))

	if withTLS {
		options = append(options, grpc.WithTransportCredentials(credentials.NewTLS(b.tlsConfig)))
	} else {
		options = append(options, grpc.WithInsecure())
	}

	return options
}
