package tracing

import (
	"io"

	"github.com/caring/go-packages/v2/pkg/logging"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-lib/metrics/prometheus"
)

// Tracer is a service object for accessing and creating tracing utils
type Tracer struct {
	tracer        opentracing.Tracer
	reporter      jaeger.Reporter
	tracingCloser io.Closer
}

// Close closes the tracing and reporting objects
func (t *Tracer) Close() error {
	t.reporter.Close()
	return t.tracingCloser.Close()
}

// GetInternalTracer returns a pointer to the internal tracer
func (t *Tracer) GetInternalTracer() *opentracing.Tracer {
	return &t.tracer
}

// NewTracer configures a jaeger tracing setup wrapped an a Tracer form this package
func NewTracer(config *Config) (*Tracer, error) {
	t := Tracer{}

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
			jaeger.NewLoggingReporter(logging.NewJaegerLogger(l)),
			jaeger.NewRemoteReporter(transport,
				jaeger.ReporterOptions.Metrics(metrics),
				jaeger.ReporterOptions.Logger(logging.NewJaegerLogger(l)),
			),
		)
	} else {
		// Simple, logging only reporter
		t.reporter = jaeger.NewLoggingReporter(logging.NewJaegerLogger(l))
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
