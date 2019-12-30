package main

import (
	"log"

	"github.com/caring/go-packages/pkg/logging"
	"github.com/caring/go-packages/pkg/tracing"
	"google.golang.org/grpc"
)

func main() {
	t, err := tracing.NewTracing(
		"my-service",
		"hostname:6451",
		false,
		false,
		// This would need to be a constructed logger in practice, a literal is used here for brevity.
		// See the logging example...
		logging.LogDetails{},
		map[string]string{
			"tag": "value",
		},
	)
	defer t.CloseTracing()

	if err != nil {
		log.Fatal("Error establishing tracing")
	}

	// Create gRPC interceptsors
	streamI := t.NewGRPCStreamServerInterceptor()
	unaryI := t.NewGRPCUnaryServerInterceptor()

	// Use your interceptors
	grpc.NewServer(
		grpc.StreamInterceptor(streamI),
		grpc.UnaryInterceptor(unaryI),
	)
}
