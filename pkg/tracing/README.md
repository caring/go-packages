#### Tracing

### About

This package contains tooling to configure tracing objects and middleware to be used between services and in tandem with gRPC. This package is a thin wrapper around [jaeger tracing go](https://github.com/jaegertracing/jaeger-client-go) which supports the open tracing standard.

### Configuration

Most of the tracing configuration can be done through environment variables. Any values passed into the tracing initialization config object will overwrite environment variables. See the table below for details.

Property| Description | Default
--- | --- | ---
SERVICE_NAME | The service name | "" Empty String
TRACE_DESTINATION_DNS | The DNS at which the trace collector is located | "" Empty String
TRACE_DESTINATION_PORT | The port of the trace collector destination | "" Empty String
TRACE_DISABLE | Boolean flag to disable trace reporting | "TRUE"
TRACE_SAMPLE_RATE | The rate spans are sampled expressed as a float. 0.8 is 80%, 0.9 is 90% etc. | "0.0"


### Usage

```golang

config := &Config{
  ServiceName: "myservice",
  Logger: logging.LogDetails{},
  GlobalTags: map[string]string{
    "my-tag": "my-value",
  },
}

  tracer, err := NewTracer(config)

  defer tracer.Close()

  // To obtain a chain interceptor for you gRPC server...
  tracer.NewGRPCUnaryServerInterceptor()
  // or
  tracer.NewGRPCStreamServerInterceptor()

```
