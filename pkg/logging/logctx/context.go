package logctx

import (
	"context"

	"github.com/caring/go-packages/v2/pkg/logging"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
)

type ctxMarker struct{}

type ctxLogger struct {
	logger *logging.Logger
}

var (
	ctxKey     = &ctxMarker{}
	nullLogger = logging.NewNopLogger()
)

// Extract gets a Logger instance from a context.Context, it always returns a logger
// populated with the latest gRPC tags
func Extract(ctx context.Context) *logging.Logger {
	l, ok := ctx.Value(ctxKey).(*ctxLogger)
	if !ok || l == nil {
		return nullLogger
	}

	fields := TagsToFields(ctx)
	return l.logger.With(nil, fields...)
}

// TagsToFields transforms the gRPC Tags on the supplied context into structured fields.
func TagsToFields(ctx context.Context) []logging.Field {
	fields := []logging.Field{}
	tags := grpc_ctxtags.Extract(ctx)
	for k, v := range tags.Values() {
		fields = append(fields, logging.Any(k, v))
	}
	return fields
}

// ToContext adds the Logger to the context for extraction later.
// Returning the new context that has been created.
func ToContext(ctx context.Context, logger *logging.Logger) context.Context {
	l := &ctxLogger{
		logger: logger,
	}
	return context.WithValue(ctx, ctxKey, l)
}
