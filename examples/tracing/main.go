package main

import (
	"log"

	"github.com/caring/go-packages/pkg/tracing"
	"google.golang.org/grpc"
)

func main() {
	b := false
	tracer, err := tracing.NewTracer(&tracing.Config{
		ServiceName:          "my-service",
		TraceDestinationDNS:  "hostname",
		TraceDestinationPort: "3000",
		DisableReporting:     &b,
		SampleRate:           0.5,
		GlobalTags: map[string]string{
			"tag": "value",
		},
	})
	defer tracer.Close()

	if err != nil {
		log.Fatal("Error establishing tracing")
	}

	// Create gRPC interceptsors
	streamI := tracer.NewGRPCStreamServerInterceptor()
	unaryI := tracer.NewGRPCUnaryServerInterceptor()

	// Use your interceptors
	grpc.NewServer(
		grpc.StreamInterceptor(streamI),
		grpc.UnaryInterceptor(unaryI),
	)
}
