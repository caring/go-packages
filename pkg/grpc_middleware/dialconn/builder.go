package dialconn

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/caring/go-packages/v2/pkg/errors"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type Builder interface {
	WithContext(ctx context.Context)
	GetContext() context.Context
	WithOptions(opts ...grpc.DialOption)
	WithUnaryInterceptors(interceptors []grpc.UnaryClientInterceptor)
	GetUnaryInterceptors() []grpc.UnaryClientInterceptor
	WithStreamInterceptors(interceptors []grpc.StreamClientInterceptor)
	GetStreamInterceptors() []grpc.StreamClientInterceptor
	WithKeepAliveParams(params keepalive.ClientParameters)
	WithClientTransportCredentials(insecureSkipVerify bool, certPool *x509.CertPool)
	GetClientTransportCredentials() credentials.TransportCredentials
	SetConnInfo(dns, port string) error
	GetConnInfo() (dns string, port string, err error)
	GetConnection(withTLS bool) (*grpc.ClientConn, error)
}

type ConnBuilder struct {
	ctx                  context.Context
	options              []grpc.DialOption
	enabledReflection    bool
	shutdownHook         func()
	enabledHealthCheck   bool
	transportCredentials credentials.TransportCredentials
	uinterceptors        []grpc.UnaryClientInterceptor
	sinterceptors        []grpc.StreamClientInterceptor
	dns                  *string
	port                 *string
	err                  error
}

// WithContext sets the context to be used
// connector will use Background if this is not set
func (b *ConnBuilder) WithContext(ctx context.Context) {
	b.ctx = ctx
}

// GetContext returns the builder context
func (b *ConnBuilder) GetContext() context.Context {
	return b.ctx
}

// WithOptions allows possing in multiple grpc dial options
// DialOption configures how we set up the connection.
func (b *ConnBuilder) WithOptions(opts ...grpc.DialOption) {
	b.options = append(b.options, opts...)
}

// WithBlock returns a DialOption which makes caller of Dial blocks until the
// underlying connection is up. Without this, Dial returns immediately and
// connecting the server happens in background.
func (b *ConnBuilder) WithBlock() {
	b.options = append(b.options, grpc.WithBlock())
}

// WithKeepAliveParams set the keep alive params
// ClientParameters is used to set keepalive parameters on the client-side.
// These configure how the client will actively probe to notice when a connection
// is broken and send pings so intermediaries will be aware of the liveness of
// the connection. Make sure these parameters are set in coordination with the
// keepalive policy on the server, as incompatible settings can result in closing
// of connection.
func (b *ConnBuilder) WithKeepAliveParams(params keepalive.ClientParameters) {
	keepAlive := grpc.WithKeepaliveParams(params)
	b.options = append(b.options, keepAlive)
}

// WithUnaryInterceptors set a list of interceptors to the Grpc client for unary
// connection. We leverage grpc_middleware package ChainUnaryClient to creates a
// single interceptor out of a chain of many interceptors to override the grpc
// default behavior of only allowing one.
func (b *ConnBuilder) WithUnaryInterceptors(interceptors []grpc.UnaryClientInterceptor) {
	b.uinterceptors = interceptors
	b.options = append(b.options, grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(interceptors...)))
}

// GetUnaryInterceptors returns the UnaryClientInterceptors slice
func (b *ConnBuilder) GetUnaryInterceptors() []grpc.UnaryClientInterceptor {
	return b.uinterceptors
}

// WithStreamInterceptors set a list of interceptors to the Grpc client for unary
// connection. We leverage grpc_middleware package ChainStreamClient to creates a
// single interceptor out of a chain of many interceptors to override the grpc
// default behavior of only allowing one.
func (b *ConnBuilder) WithStreamInterceptors(interceptors []grpc.StreamClientInterceptor) {
	b.sinterceptors = interceptors
	b.options = append(b.options, grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(interceptors...)))
}

// GetStreamInterceptors returns the StreamClientInterceptors slice
func (b *ConnBuilder) GetStreamInterceptors() []grpc.StreamClientInterceptor {
	return b.sinterceptors
}

// WithClientTransportCredentials builds transport credentials for a gRPC client
func (b *ConnBuilder) WithClientTransportCredentials(insecureSkipVerify bool, certPool *x509.CertPool) {
	var tlsConf tls.Config

	if insecureSkipVerify {
		tlsConf.InsecureSkipVerify = true
		b.transportCredentials = credentials.NewTLS(&tlsConf)
		return
	}

	tlsConf.RootCAs = certPool
	b.transportCredentials = credentials.NewTLS(&tlsConf)
}

func (b *ConnBuilder) GetClientTransportCredentials() credentials.TransportCredentials {
	return b.transportCredentials
}

// SetConnInfo allows passing in the dns and port for the connection, providing
// flexibility if a consumer wants to set a default and or override
func (b *ConnBuilder) SetConnInfo(dns, port string) error {

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
	if err != nil || i < 0 || i > 65353 {
		return errors.Wrap(err, fmt.Sprintf("invalid port number: %s", port))
	}
	b.port = &port

	return nil

}

// GetConnInfo returns the dns and port that were set, and errors if either
// were null, allowing verfication prior to attempting the connection
func (b *ConnBuilder) GetConnInfo() (dns string, port string, err error) {
	if b.dns == nil || b.port == nil {
		return "", "", errors.New("Connection info not fully set")
	}
	return *b.dns, *b.port, nil
}

// GetConnection returns the client connection to the server
// withTLS will return a TLS encryped connection
func (b *ConnBuilder) GetConnection(withTLS bool) (*grpc.ClientConn, error) {

	dns, port, err := b.GetConnInfo()
	if err != nil {
		return nil, fmt.Errorf("target connection parameter missing: dns and/or port not set.")
	}

	addr := dns + ":" + port
	log.Printf("Target to connect = %s", addr)

	if withTLS == true {
		b.options = append(b.options, grpc.WithTransportCredentials(b.transportCredentials))
	} else {
		b.options = append(b.options, grpc.WithInsecure())
	}
	cc, err := grpc.DialContext(b.getContext(), addr, b.options...)

	if err != nil {
		return nil, fmt.Errorf("unable to connect to client. error = %+v", err)
	}
	return cc, nil
}

// returns the builder context, or uses background if not set
func (b *ConnBuilder) getContext() context.Context {
	ctx := b.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx
}
