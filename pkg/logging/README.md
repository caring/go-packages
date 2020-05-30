#### Tracing

### About

This package contains tooling to configure and create fast, structured loggers. This package is a thin wrapper around [Uber Zap](https://github.com/uber-go/zap).

### Configuration

Most of the logging configuration can be done through environment variables. Any values passed into the logging initialization config object will overwrite environment variables. See the table below for details. AWS authorization, region and all other account info are pulled from the hardware that the logger is running on.

Property | Description | Default
--- | --- | ---
SERVICE_NAME | The service name | "" Empty String
LOG_NAME | The name of the logger | "" Empty String
LOG_LEVEL | The lowest logged level. All all levels above this will be logged to all enabled outputs | "INFO"
LOG_ENABLE_DEV | Boolean which enables the developer log configuration compatible with zap-pretty | "FALSE"
LOG_STREAM_MONITORING | The name of the kinesis stream where developer monitoring logs are piped through | "" Empty String
LOG_STREAM_REPORTING | The name of the kinesis stream where business insight lgs are piped through | "" Empty String
LOG_DISABLE_KINESIS | Boolean flag to disable out put to kinesis, generally only enabled in Prod | "TRUE"


### Usage

```golang

t := true
f := false
config := &Config{
		LoggerName:          "my-logger",
		ServiceName:         "my-service",
		LogLevel:            "INFO",
		EnableDevLogging:    &f,
		KinesisStreamName:   "stream-1",
		KinesisPartitionKey: "shard-1",
		DisableKinesis:      &t,
  },
}

  logger, err := NewLogger(config)

  logger.Warn("sample message", logging.Int64("fieldA", 3))

  // To obtain a chain interceptor for you gRPC server...
  logger.NewGRPCUnaryServerInterceptor()
  // or
  logger.NewGRPCStreamServerInterceptor()

```

### Pretty Printing

The development logger outputs logs in a format usable by [this tool](https://github.com/maoueh/zap-pretty).

First brew install it
```bash
$ brew install maoueh/tap/zap-pretty
```

Then when you run your service pipe the output to the pretty-printer
```bash
$ ./main | zap-pretty
```
