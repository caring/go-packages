#### Tracing

### About

This package contains tooling to configure tracing objects and middleware to be used between services and in tandem with gRPC. This package is a thin wrapper around [jaeger tracing go](https://github.com/jaegertracing/jaeger-client-go) which supports the open tracing standard.

### Configuration

Most of the tracing configuration can be done through environment variables. Any values passed into the tracing initialization config object will overwrite environment variables. See the table below for details.

Property| Description
--- | ---
SERVICE_NAME | The service name
TRACE_DESTINATION_DNS | The DNS at which the trace collector is located
TRACE_DESTINATION_PORT | The port of the trace collector destination
TRACE_DISABLE | Setting this value to true disables trace reporting
TRACE_SAMPLE_RATE | The rate spans are sampled expressed as a float. 0.8 is %80, 0.9 is 90% etc.


### Usage

```golang
config := &TracerConfig{
  ServiceName: "myservice",
  Logger: logging.LogDetails,
  GlobalTags: map[string]string{
    "my-tag": "my-value",
  }

  tracer, err := NewTracer(config)

  defer tracer.Close()

  // To obtain a chain interceptor for you gRPC server...
  tracer.NewGRPCUnaryServerInterceptor()
  // or
  tracer.NewGRPCStreamServerInterceptor()
}
```
