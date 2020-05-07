package tracing

import (
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"google.golang.org/grpc"
)

// Tracer provides an interface by which to interact with the tracing objects created by this package
type Tracer interface {
	// CloseTracing closes the tracing and reporting objects that
	// are constructed within the tracing package
	Close() error
	// GetInternalTracer returns a pointer to the internal tracer.
	//
	// Note, The internal tracing package may change.
	GetInternalTracer() *opentracing.Tracer
	// NewGRPCUnaryServerInterceptor returns a gRPC interceptor wrapped around the internal tracer
	NewGRPCUnaryServerInterceptor() grpc.UnaryServerInterceptor
	// NewGRPCStreamServerInterceptor returns a gRPC stream interceptor wrapped around the internal tracer
	NewGRPCStreamServerInterceptor() grpc.StreamServerInterceptor
}

// tracerImpl implements the Tracing interface using the jaeger tracing package
type tracerImpl struct {
	tracer        opentracing.Tracer
	reporter      jaeger.Reporter
	tracingCloser io.Closer
}

// Close closes the tracing and reporting objects
func (t *tracerImpl) Close() error {
	t.reporter.Close()
	return t.tracingCloser.Close()
}

// GetInternalTracer returns a pointer to the internal tracer
func (t *tracerImpl) GetInternalTracer() *opentracing.Tracer {
	return &t.tracer
}

// NewTracer configures a jaeger tracing setup and returns the the configured tracer and reporter for use
func NewTracer(config *Config) (Tracer, error) {
	t := tracerImpl{}

	c, err := mergeAndPopulateConfig(config)
	if err != nil {
		return nil, err
	}

	factory := prometheus.New()
	metrics := jaeger.NewMetrics(factory, c.GlobalTags)

	l := c.Logger

	if !*c.DisableReporting {
		transport, err := jaeger.NewUDPTransport(c.TraceDestinationDNS+":"+c.TraceDestinationPort, 0)
		if err != nil {
			return nil, err
		}

		// create composite logger to log to the logger and report to the
		// remote server
		t.reporter = jaeger.NewCompositeReporter(
			jaeger.NewLoggingReporter(l.NewJaegerLogger()),
			jaeger.NewRemoteReporter(transport,
				jaeger.ReporterOptions.Metrics(metrics),
				jaeger.ReporterOptions.Logger(l.NewJaegerLogger()),
			),
		)
	} else {
		// Simple, logging only reporter
		t.reporter = jaeger.NewLoggingReporter(l.NewJaegerLogger())
	}

	// create a sampler for the spans so that we don't report every single span which would be untenable
	sampler, err := jaeger.NewGuaranteedThroughputProbabilisticSampler(1.0, c.SampleRate)
	if err != nil {
		return nil, err
	}

	// now make the tracer
	t.tracer, t.tracingCloser = jaeger.NewTracer(
		c.ServiceName,
		sampler,
		t.reporter,
		jaeger.TracerOptions.Metrics(metrics),
	)

	opentracing.SetGlobalTracer(t.tracer)

	return &t, nil
}
