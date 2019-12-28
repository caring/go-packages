package tracing

import (
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaeger_zap "github.com/uber/jaeger-client-go/log/zap"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"
)

// Tracing contains all of the tracing object and methods needed to interact with
// the constructed tracing interface
type Tracing struct {
	Tracer        opentracing.Tracer
	reporter      jaeger.Reporter
	tracingCloser io.Closer
}

// CloseTracing closes the tracing and reporting objects
func (t *Tracing) CloseTracing() error {
	t.reporter.Close()
	return t.tracingCloser.Close()
}

// NewTracer configures a jaeger tracing setup and returns the the configured tracer and reporter for use
//
// Arguments:
// serviceName: The name of the service (app) in tracing messages
// transportDestination: The jaeger server where we are sending the remote reporting to if enabled. <host>:<port> ie "localhost:3001"
// reportRemote: True enables remote reporting
// logger accepts: Zap logger to use for logging tracing reporting
// metricTags: Key value tags appended to the tracing logs
// lowerBound: The guaranteed minimum amount samples per endpoint per timeframe. See jaeger client docs https://github.com/jaegertracing/jaeger-client-go/blob/master/sampler.go#L241
// sampleRate: The percentage of samples to report expressed as a float between 0.0 and 1.0
//
func NewTracer(serviceName, transportDestination string, reportRemote bool, logger *zap.Logger, metricTags map[string]string, lowerBound, sampleRate float64) (*Tracing, error) {
	t := Tracing{}

	// Adapt the zap logger to work with jaeger
	adaptedLogger := jaeger_zap.NewLogger(logger)

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
			jaeger.NewLoggingReporter(adaptedLogger),
			jaeger.NewRemoteReporter(transport,
				jaeger.ReporterOptions.Metrics(metrics),
				jaeger.ReporterOptions.Logger(adaptedLogger),
			),
		)
	} else {
		// Simple, logging only reporter
		t.reporter = jaeger.NewLoggingReporter(adaptedLogger)
	}

	// create a sampler for the spans so that we don't report every single span which would be untenable
	sampler, err := jaeger.NewGuaranteedThroughputProbabilisticSampler(lowerBound, sampleRate)
	if err != nil {
		return nil, err
	}

	// now make the tracer
	t.Tracer, t.tracingCloser = jaeger.NewTracer(
		serviceName,
		sampler,
		t.reporter,
		jaeger.TracerOptions.Metrics(metrics),
	)

	opentracing.SetGlobalTracer(t.Tracer)

	return &t, nil
}
