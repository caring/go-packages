package logging

import (
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/uber/jaeger-client-go"
	jaeger_zap "github.com/uber/jaeger-client-go/log/zap"
	"google.golang.org/grpc"
)

// NewJaegerLogger returns a jaeger logging interface implementer that has been populated
// with Loggers internal and accumulated fields as well as settings
func (l *Logger) NewJaegerLogger() jaeger.Logger {
	populatedL := l.GetInternalLogger().With(l.getZapFields()...)
	j := jaeger_zap.NewLogger(populatedL)

	return j
}

// NewJaegerLogger returns a jaeger logging interface implementer that has been populated
// with Loggers internal and accumulated fields as well as settings
func NewJaegerLogger(l Logging) jaeger.Logger {
	populatedL := l.GetInternalLogger().With(l.getZapFields()...)
	j := jaeger_zap.NewLogger(populatedL)

	return j
}

// NewGRPCUnaryServerInterceptor returns a gRPC unary interceptor that has been populated
// with Loggers internal and accumulated fields as well as settings
func (l *Logger) NewGRPCUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	populatedL := l.GetInternalLogger().With(l.getZapFields()...)

	return grpc_zap.UnaryServerInterceptor(populatedL)
}

func NewGRPCUnaryServerInterceptor(l Logging) grpc.UnaryServerInterceptor {
	populatedL := l.GetInternalLogger().With(l.getZapFields()...)

	return grpc_zap.UnaryServerInterceptor(populatedL)
}

// NewGRPCStreamServerInterceptor returns a gRPC stream interceptor that has been populated
// with Loggers internal and accumulated fields as well as settings
func (l *Logger) NewGRPCStreamServerInterceptor() grpc.StreamServerInterceptor {
	populatedL := l.GetInternalLogger().With(l.getZapFields()...)

	return grpc_zap.StreamServerInterceptor(populatedL)
}

func NewGRPCStreamServerInterceptor(l Logging) grpc.StreamServerInterceptor {
	populatedL := l.GetInternalLogger().With(l.getZapFields()...)

	return grpc_zap.StreamServerInterceptor(populatedL)
}
