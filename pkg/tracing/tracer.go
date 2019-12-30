package tracing

import (
	"io"

	"github.com/caring/go-packages/pkg/logging"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"google.golang.org/grpc"
)

// Tracing provides an interface by which to interact with the tracing objects created by this package
type Tracing interface {
	CloseTracing() error
	GetInternalTracer() *opentracing.Tracer
	NewGRPCUnaryServerInterceptor() grpc.UnaryServerInterceptor
	NewGRPCStreamServerInterceptor() grpc.StreamServerInterceptor
}

// JTracing implements the Tracing interface using the jaeger tracing package
type jTracing struct {
	tracer        opentracing.Tracer
	reporter      jaeger.Reporter
	tracingCloser io.Closer
}

// CloseTracing closes the tracing and reporting objects
func (t *jTracing) CloseTracing() error {
	t.reporter.Close()
	return t.tracingCloser.Close()
}

// GetInternalTracer returns a pointer to the internal tracer
func (t *jTracing) GetInternalTracer() *opentracing.Tracer {
	return &t.tracer
}

// NewGRPCUnaryServerInterceptor returns a gRPC interceptor wrapped around the internal tracer
func (t *jTracing) NewGRPCUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(t.tracer))
}

// NewGRPCStreamServerInterceptor returns a gRPC stream interceptor wrapped around the internal tracer
func (t *jTracing) NewGRPCStreamServerInterceptor() grpc.StreamServerInterceptor {
	return grpc_opentracing.StreamServerInterceptor(grpc_opentracing.WithTracer(t.tracer))
}

// NewTracing configures a jaeger tracing setup and returns the the configured tracer and reporter for use
//
// Arguments:
// serviceName: The name of the service (app) in tracing messages
// transportDestination: The jaeger server where we are sending the remote reporting to if enabled. <host>:<port> ie "localhost:3001"
// reportRemote: True enables remote reporting
// logger: accepts the caring logger to use for logging tracing reporting
// metricTags: Key value tags appended to the tracing logs
// lowerBound: The guaranteed minimum amount samples per endpoint per timeframe. See jaeger client docs https://github.com/jaegertracing/jaeger-client-go/blob/master/sampler.go#L241
// sampleRate: The percentage of samples to report expressed as a float between 0.0 and 1.0
//
func NewTracing(serviceName, transportDestination string, reportRemote bool, logger logging.LogDetails, metricTags map[string]string, lowerBound, sampleRate float64) (Tracing, error) {
	t := jTracing{}

	// create a metrics object
	factory := prometheus.New()
	metrics := jaeger.NewMetrics(factory, metricTags)

	if reportRemote {
		// If we want to report to a remote tracing analytics server
		// create the connection here
		transport, err := jaeger.NewUDPTransport(transportDestination, 0)
		if err != nil {
			return nil, err
		}

		// create composite logger to log to the logger and report to the
		// remote server
		t.reporter = jaeger.NewCompositeReporter(
			jaeger.NewLoggingReporter(logger.NewJaegerLogger()),
			jaeger.NewRemoteReporter(transport,
				jaeger.ReporterOptions.Metrics(metrics),
				jaeger.ReporterOptions.Logger(logger.NewJaegerLogger()),
			),
		)
	} else {
		// Simple, logging only reporter
		t.reporter = jaeger.NewLoggingReporter(logger.NewJaegerLogger())
	}

	// create a sampler for the spans so that we don't report every single span which would be untenable
	sampler, err := jaeger.NewGuaranteedThroughputProbabilisticSampler(lowerBound, sampleRate)
	if err != nil {
		return nil, err
	}

	// now make the tracer
	t.tracer, t.tracingCloser = jaeger.NewTracer(
		serviceName,
		sampler,
		t.reporter,
		jaeger.TracerOptions.Metrics(metrics),
	)

	opentracing.SetGlobalTracer(t.tracer)

	return &t, nil
}
