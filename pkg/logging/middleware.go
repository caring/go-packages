package logging

import (
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/uber/jaeger-client-go"
	jaeger_zap "github.com/uber/jaeger-client-go/log/zap"
	"google.golang.org/grpc"
)

// NewJaegerLogger creates a logger that implements the jaeger logger interface
// and is populated by both the loggers parent fields and the log details provided
func (l *loggerImpl) NewJaegerLogger() jaeger.Logger {
	populatedL := l.internalLogger.With(l.getZapFields()...)
	j := jaeger_zap.NewLogger(populatedL)

	return j
}

// NewGRPCUnaryServerInterceptor creates a gRPC unary interceptor that is wrapped around
// the internal logger populated with its parents fields and any provided log details
func (l *loggerImpl) NewGRPCUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	populatedL := l.internalLogger.With(l.getZapFields()...)

	return grpc_zap.UnaryServerInterceptor(populatedL)
}

// NewGRPCStreamServerInterceptor creates a gRPC stream interceptor that is wrapped around
// the internal logger populated with its parents fields and any provided log details
func (l *loggerImpl) NewGRPCStreamServerInterceptor() grpc.StreamServerInterceptor {
	populatedL := l.internalLogger.With(l.getZapFields()...)

	return grpc_zap.StreamServerInterceptor(populatedL)
}
