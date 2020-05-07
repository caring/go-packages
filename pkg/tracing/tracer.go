package tracing

import (
	"io"
	"os"
	"strconv"

	"github.com/caring/go-packages/pkg/logging"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
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

// NewGRPCUnaryServerInterceptor returns a gRPC interceptor wrapped around the internal tracer
func (t *tracerImpl) NewGRPCUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(t.tracer))
}

// NewGRPCStreamServerInterceptor returns a gRPC stream interceptor wrapped around the internal tracer
func (t *tracerImpl) NewGRPCStreamServerInterceptor() grpc.StreamServerInterceptor {
	return grpc_opentracing.StreamServerInterceptor(grpc_opentracing.WithTracer(t.tracer))
}

// TracerConfig contains initialization config for NewTracer
type TracerConfig struct {
	// The name of the service this tracer is being used in
	ServiceName string
	// The DNS of the tracing collector which traces are reported to.
	TraceDestinationDNS string
	// The port of the tracing collector which traces are reported to.
	TraceDestinationPort string
	// Boolean to disable sending tracing reports
	DisableReporting *bool
	// Our Tracing setup uses jaegers GuaranteedThroughputProbabilisticSampler.
	// This number determins what percent of our traces are sampled. 0.8 = %80, 0.9 = 90% etc.
	// See their docs on sampling https://github.com/jaegertracing/jaeger-client-go#sampling
	// Or the source code for this sampler https://github.com/jaegertracing/jaeger-client-go/blob/master/sampler.go#L242
	SampleRate float64
	// The instance of our own logger to use for logging traces
	Logger logging.Logger
	// key values pairs that will be included on all spans
	GlobalTags map[string]string
}

// NewTracer configures a jaeger tracing setup and returns the the configured tracer and reporter for use
func NewTracer(config *TracerConfig) (Tracer, error) {
	t := tracerImpl{}

	populatedConfig, err := getEnvConfig(config)
	if err != nil {
		return nil, err
	}

	factory := prometheus.New()
	metrics := jaeger.NewMetrics(factory, populatedConfig.GlobalTags)

	l := populatedConfig.Logger

	if !*populatedConfig.DisableReporting {
		transport, err := jaeger.NewUDPTransport(populatedConfig.TraceDestinationDNS+":"+populatedConfig.TraceDestinationPort, 0)
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
	sampler, err := jaeger.NewGuaranteedThroughputProbabilisticSampler(1.0, populatedConfig.SampleRate)
	if err != nil {
		return nil, err
	}

	// now make the tracer
	t.tracer, t.tracingCloser = jaeger.NewTracer(
		populatedConfig.ServiceName,
		sampler,
		t.reporter,
		jaeger.TracerOptions.Metrics(metrics),
	)

	opentracing.SetGlobalTracer(t.tracer)

	return &t, nil
}

// getEnvConfig populates a config object from the environment, if any of the values are 0 values
func getEnvConfig(config *TracerConfig) (*TracerConfig, error) {
	final := TracerConfig{}
	if config.ServiceName == "" {
		final.ServiceName = os.Getenv("SERVICE_NAME")
	} else {
		final.ServiceName = config.ServiceName
	}

	if config.TraceDestinationDNS == "" {
		final.TraceDestinationDNS = os.Getenv("TRACE_DESTINATION_DNS")
	} else {
		final.TraceDestinationDNS = config.TraceDestinationDNS
	}

	if config.TraceDestinationPort == "" {
		final.TraceDestinationDNS = os.Getenv("TRACE_DESTINATION_PORT")
	} else {
		final.TraceDestinationDNS = config.TraceDestinationDNS
	}

	if config.DisableReporting == nil {
		boolString := os.Getenv("TRACE_DISABLE")
		v, err := strconv.ParseBool(boolString)
		if err != nil {
			return nil, err
		}
		final.DisableReporting = &v
	} else {
		final.DisableReporting = config.DisableReporting
	}

	if config.SampleRate == 0 {
		floatString := os.Getenv("TRACE_SAMPLE_RATE")
		v, err := strconv.ParseFloat(floatString, 64)
		if err != nil {
			return nil, err
		}
		final.SampleRate = v
	} else {
		final.SampleRate = config.SampleRate
	}

	if config.GlobalTags == nil {
		final.GlobalTags = map[string]string{}
	} else {
		final.GlobalTags = config.GlobalTags
	}

	return &final, nil
}
