package main

import (
	"log"

	"github.com/caring/go-packages/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

func main() {
	t, err := tracing.NewTracer(
		"my-service",
		"hostname:6451",
		false,
		zap.L(),
		map[string]string{
			"tag": "value",
		},
		1.0,
		1.0,
	)
	defer tracing.CloseTracing()

	if err != nil {
		log.Fatal("Error establishing tracing")
	}

	// use the tracer someplace
	opentracing.SetGlobalTracer(t.Tracer)
}
