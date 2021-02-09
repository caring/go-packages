package grpc_middleware

import (
	"github.com/caring/go-packages/v2/pkg/logging"
	"github.com/caring/go-packages/v2/pkg/tracing"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// StreamOptions wraps the input for stream interceptor chain creation
type StreamOptions struct {
	Logger       *logging.Logger
	Tracer       *tracing.Tracer
	Interceptors []grpc.StreamServerInterceptor
}

// NewGRPCChainedStreamInterceptor creates new stream interceptors from each package in this library
// that gets passed in to the options block, then chains them in a consistent order into a single stream
// interceptor. You may pass additional interceptors to be added to the *end* of the chain. They will be
// added in the order they are passed in.
//
// Interceptor ordering is important, context is passed from interceptor to interceptor in the
// order they are passed to the chain. In most cases, logging and tracing will go first, because they are
// fundamental to reporting every output and process in our services.
//
// If you need more control over the ordering of chained interceptors, you may build your own chain by retreiving
// individual interceptors from each package in this library.
func NewGRPCChainedStreamInterceptor(opts StreamOptions) grpc.ServerOption {
	chain := []grpc.StreamServerInterceptor{}

	if opts.Logger != nil {
		chain = append(chain, opts.Logger.NewGRPCStreamServerInterceptor())
	}
	if opts.Tracer != nil {
		chain = append(chain, opts.Tracer.NewGRPCStreamServerInterceptor())
	}
	if opts.Interceptors != nil {
		chain = append(chain, opts.Interceptors...)
	}

	return grpc.StreamInterceptor(
		grpc_middleware.ChainStreamServer(
			chain...,
		),
	)
}

// UnaryOptions wraps the input for unary interceptor chain creation
type UnaryOptions struct {
	Logger       *logging.Logger
	Tracer       *tracing.Tracer
	Interceptors []grpc.UnaryServerInterceptor
}

// NewGRPCChainedUnaryInterceptor creates new unary interceptors from each package in this library
// that gets passed in to the options block, then chains them in a consistent order into a single unary
// interceptor. You may pass additional interceptors to be added to the *end* of the chain. They will be
// added in the order they are passed in.
//
// Interceptor ordering is important, context is passed from interceptor to interceptor in the
// order they are passed to the chain. In most cases, logging and tracing will go first, because they are
// fundamental to reporting every output and process in our services.
//
// If you need more control over the ordering of chained interceptors, you may build your own chain by retreiving
// individual interceptors from each package in this library.
func NewGRPCChainedUnaryInterceptor(opts UnaryOptions) grpc.ServerOption {
	chain := []grpc.UnaryServerInterceptor{}

	if opts.Logger != nil {
		chain = append(chain, opts.Logger.NewGRPCUnaryServerInterceptor())
	}
	if opts.Tracer != nil {
		chain = append(chain, opts.Tracer.NewGRPCUnaryServerInterceptor())
	}
	if opts.Interceptors != nil {
		chain = append(chain, opts.Interceptors...)
	}

	return grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(
			chain...,
		),
	)
}
