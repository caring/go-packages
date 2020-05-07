package tracing

import (
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"google.golang.org/grpc"
)

// NewGRPCUnaryServerInterceptor returns a gRPC interceptor wrapped around the internal tracer
func (t *tracerImpl) NewGRPCUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(t.tracer))
}

// NewGRPCStreamServerInterceptor returns a gRPC stream interceptor wrapped around the internal tracer
func (t *tracerImpl) NewGRPCStreamServerInterceptor() grpc.StreamServerInterceptor {
	return grpc_opentracing.StreamServerInterceptor(grpc_opentracing.WithTracer(t.tracer))
}
