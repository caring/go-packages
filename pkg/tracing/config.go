package tracing

import (
	"os"
	"strconv"

	"github.com/caring/go-packages/pkg/logging"
)

// Config contains initialization config for NewTracer
type Config struct {
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
	Logger *logging.Logger
	// key values pairs that will be included on all spans
	GlobalTags map[string]string
}

var (
	trueVar  = true
	falseVar = false
)

func newDefaultConfig() *Config {
	return &Config{
		ServiceName:          "",
		TraceDestinationDNS:  "",
		TraceDestinationPort: "",
		DisableReporting:     &trueVar,
		SampleRate:           0.0,
		Logger:               nil,
		GlobalTags:           nil,
	}
}

// mergeAndPopulateConfig starts with a default config, and populates
// it with config from the environment. Config from the environment can
// be overridden with any config input as arguments. Only non 0 values will
// overwrite the defaults
func mergeAndPopulateConfig(c *Config) (*Config, error) {
	final := newDefaultConfig()

	if c.ServiceName != "" {
		final.ServiceName = c.ServiceName
	} else if s := os.Getenv("SERVICE_NAME"); s != "" {
		final.ServiceName = s
	}

	if c.TraceDestinationDNS != "" {
		final.TraceDestinationDNS = c.TraceDestinationDNS
	} else if s := os.Getenv("TRACE_DESTINATION_DNS"); s != "" {
		final.TraceDestinationDNS = s
	}

	if c.TraceDestinationPort != "" {
		final.TraceDestinationPort = c.TraceDestinationPort
	} else if s := os.Getenv("TRACE_DESTINATION_PORT"); s != "" {
		final.TraceDestinationPort = s
	}

	if c.DisableReporting != nil {
		final.DisableReporting = c.DisableReporting
	} else if s := os.Getenv("TRACE_DISABLE"); s != "" {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return nil, err
		}
		final.DisableReporting = &b
	}

	if c.SampleRate != 0 {
		final.SampleRate = c.SampleRate
	} else if s := os.Getenv("TRACE_SAMPLE_RATE"); s != "" {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
		final.SampleRate = v
	}

	if c.GlobalTags != nil {
		final.GlobalTags = c.GlobalTags
	} else {
		final.GlobalTags = map[string]string{}
	}

	return final, nil
}
