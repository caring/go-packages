#### About

This package is a utils package that helps guarantee consistency and ease for inserting the other packages in this library as interceptors into gRPC unary and stream calls.

### Usage

```golang
  l := &logging.Logger{}
  t := &tracing.Tracer{}

  streamOpts := StreamOptions{
    Logger: l,
    Tracer: t,
  }

  unaryOpts := UnaryOptions{
    Logger: l,
    Tracer: t,
  }

  // create protocol server with chained interceptors
  g := grpc.NewServer(
    NewGRPCChainedStreamInterceptor(streamOts),
    NewGRPCChainedUnaryInterceptor(unaryOpts),
  )
```
