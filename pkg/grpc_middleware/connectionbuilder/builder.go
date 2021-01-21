package connectionbuilder

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"

	"github.com/caring/go-packages/pkg/errors"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type ConnectionBuilder interface {
	WithContext(ctx context.Context)
	WithOptions(opts ...grpc.DialOption)
	WithInsecure()
	WithUnaryInterceptors(interceptors []grpc.UnaryClientInterceptor)
	WithStreamInterceptors(interceptors []grpc.StreamClientInterceptor)
	WithKeepAliveParams(params keepalive.ClientParameters)
	SetConnInfo(dns, port string)
	GetConnInfo() (dns string, port string, err error)
	GetConnection() (*grpc.ClientConn, error)
}

type Builder struct {
	ctx                  context.Context
	options              []grpc.DialOption
	enabledReflection    bool
	shutdownHook         func()
	enabledHealthCheck   bool
	transportCredentials credentials.TransportCredentials
	dns                  *string
	port                 *string
	err                  error
}

// WithContext sets the context to be used
// connector will use Background if this is not set
func (b *Builder) WithContext(ctx context.Context) {
	b.ctx = ctx
}

// WithOptions allows possing in multiple grpc dial options
// DialOption configures how we set up the connection.
func (b *Builder) WithOptions(opts ...grpc.DialOption) {
	b.options = append(b.options, opts...)
}

// WithInsecure returns a DialOption which disables transport security for this
// ClientConn. Note that transport security is required unless WithInsecure is
// set.
func (b *Builder) WithInsecure() {
	b.options = append(b.options, grpc.WithInsecure())
}

// WithBlock returns a DialOption which makes caller of Dial blocks until the
// underlying connection is up. Without this, Dial returns immediately and
// connecting the server happens in background.
func (b *Builder) WithBlock() {
	b.options = append(b.options, grpc.WithBlock())
}

// WithKeepAliveParams set the keep alive params
// ClientParameters is used to set keepalive parameters on the client-side.
// These configure how the client will actively probe to notice when a connection
// is broken and send pings so intermediaries will be aware of the liveness of
// the connection. Make sure these parameters are set in coordination with the
// keepalive policy on the server, as incompatible settings can result in closing
// of connection.
func (b *Builder) WithKeepAliveParams(params keepalive.ClientParameters) {
	keepAlive := grpc.WithKeepaliveParams(params)
	b.options = append(b.options, keepAlive)
}

// WithUnaryInterceptors set a list of interceptors to the Grpc client for unary
// connection. We leverage grpc_middleware package ChainUnaryClient to creates a
// single interceptor out of a chain of many interceptors to override the grpc
// default behavior of only allowing one.
func (b *Builder) WithUnaryInterceptors(interceptors []grpc.UnaryClientInterceptor) {
	b.options = append(b.options, grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(interceptors...)))
}

// WithStreamInterceptors set a list of interceptors to the Grpc client for unary
// connection. We leverage grpc_middleware package ChainStreamClient to creates a
// single interceptor out of a chain of many interceptors to override the grpc
// default behavior of only allowing one.
func (b *Builder) WithStreamInterceptors(interceptors []grpc.StreamClientInterceptor) {
	b.options = append(b.options, grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(interceptors...)))
}

// WithClientTransportCredentials builds transport credentials for a gRPC client
func (b *Builder) WithClientTransportCredentials(insecureSkipVerify bool, certPool *x509.CertPool) {
	var tlsConf tls.Config

	if insecureSkipVerify {
		tlsConf.InsecureSkipVerify = true
		b.transportCredentials = credentials.NewTLS(&tlsConf)
		return
	}

	tlsConf.RootCAs = certPool
	b.transportCredentials = credentials.NewTLS(&tlsConf)
}

// SetConnInfo allows passing in the dns and port for the connection, providing
// flexibility if a consumer wants to set a default and or override
func (b *Builder) SetConnInfo(dns, port string) {
	b.dns = &dns
	b.port = &port
}

// GetConnInfo returns the dns and port that were set, and errors if either
// were null, allowing verfication prior to attempting the connection
func (b *Builder) GetConnInfo() (dns string, port string, err error) {
	if b.dns == nil || b.port == nil {
		return "", "", errors.New("Connection info not fully set")
	}
	return *b.dns, *b.port, nil
}

// GetConnection returns the client connection to the server
// withTLS will return a TLS encryped connection
func (b *Builder) GetConnection(withTLS bool) (*grpc.ClientConn, error) {

	dns, port, err := b.GetConnInfo()
	if err != nil {
		return nil, fmt.Errorf("target connection parameter missing: dns and/or port not set.")
	}

	addr := dns + ":" + port
	log.Printf("Target to connect = %s", addr)

	if withTLS == true {
		b.options = append(b.options, grpc.WithTransportCredentials(b.transportCredentials))
	}
	cc, err := grpc.DialContext(b.getContext(), addr, b.options...)

	if err != nil {
		return nil, fmt.Errorf("unable to connect to client. error = %+v", err)
	}
	return cc, nil
}

// returns the builder context, or uses background if not set
func (b *Builder) getContext() context.Context {
	ctx := b.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx
}
