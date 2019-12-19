package main

import (
	"github.com/caring/go-packages/pkg/logging"
)

func main() {
	l, err := logging.InitLogger(false, "logger-1", "my-kinesis-stream", "aws-access-key", "aws-secret-key", "aws-region")
	
	if err != nil {
		panic(err)
	}
	
	l.Warn(logging.LogDetails{
		Message:         "failed to do some",
		ServiceId:       "sample-service",
		CorrelationalId: "sample-correlation",
		TraceabilityId:  "trace-id",
		ClientId:        "client-id",
		UserId:          "userID",
		Endpoint:        "myurl.com@methodName",
		AdditionalData: map[string]string{"additional-content": "value-1"},
	})

	l.Error(logging.LogDetails{
		Message:         "failed to do some",
		ServiceId:       "sample-service",
		CorrelationalId: "sample-correlation",
		TraceabilityId:  "trace-id",
		ClientId:        "client-id",
		UserId:          "userID",
		Endpoint:        "myurl.com@methodName",
		AdditionalData: map[string]string{"additional-content": "value-1"},
	})

	l.Fatal(logging.LogDetails{
		Message:         "failed to do some",
		ServiceId:       "sample-service",
		CorrelationalId: "sample-correlation",
		TraceabilityId:  "trace-id",
		ClientId:        "client-id",
		UserId:          "userID",
		Endpoint:        "myurl.com@methodName",
		AdditionalData: map[string]string{"additional-content": "value-1"},
	})

	l.Panic(logging.LogDetails{
		Message:         "failed to do some",
		ServiceId:       "sample-service",
		CorrelationalId: "sample-correlation",
		TraceabilityId:  "trace-id",
		ClientId:        "client-id",
		UserId:          "userID",
		Endpoint:        "myurl.com@methodName",
		AdditionalData: map[string]string{"additional-content": "value-1"},
	})

	l.Info(logging.LogDetails{
		Message:         "failed to do some",
		ServiceId:       "sample-service",
		CorrelationalId: "sample-correlation",
		TraceabilityId:  "trace-id",
		ClientId:        "client-id",
		UserId:          "userID",
		Endpoint:        "myurl.com@methodName",
		AdditionalData: map[string]string{"additional-content": "value-1"},
	})

	l.Debug(logging.LogDetails{
		Message:         "failed to do some",
		ServiceId:       "sample-service",
		CorrelationalId: "sample-correlation",
		TraceabilityId:  "trace-id",
		ClientId:        "client-id",
		UserId:          "userID",
		Endpoint:        "myurl.com@methodName",
		AdditionalData: map[string]string{"additional-content": "value-1"},
	})
}
