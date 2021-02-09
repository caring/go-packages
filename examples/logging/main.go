package main

import "github.com/caring/go-packages/v2/pkg/logging"

func main() {
	t := true
	parent, err := logging.NewLogger(&logging.Config{
		LoggerName:              "logger-1",
		ServiceName:             "call-scoring",
		LogLevel:                logging.DebugLevel,
		EnableDevLogging:        &t,
		KinesisStreamMonitoring: "stream-1-monitoring",
		KinesisStreamReporting:  "stream-1-reporting",
		DisableKinesis:          &t,
	})

	if err != nil {
		panic(err)
	}

	child := parent.NewChild(
		&logging.FieldOpts{
			Endpoint: "Some-endpoint",
		},
		logging.String("child", "this wont be in the parent"),
	)

	child.Warn("sample message", logging.Int64("fieldA", 3))

	child.With(nil, logging.Bool("fieldB", true), logging.String("fieldB", "helloworld"))

	parent.Warn("here's another waring, see no child field!")

	child.With(&logging.FieldOpts{
		CorrelationID: "someID",
	})

	child.Warn("I'm the child, final warning... see my fields have changed!", logging.Float64s("floats", []float64{1.0, 2.0, 3.0}))

	m := map[string]interface{}{
		"key1": 1.0,
		"key2": true,
		"key3": "string",
		"key4": map[string]interface{}{
			"key5": 459876,
		},
	}
	obj := map[string]interface{}{
		"key1": 2134,
		"key2": true,
	}
	a := []map[string]interface{}{obj, obj}

	parent.Warn("Look how any field can marshal complex objects, this is expensive!", logging.Any("map", m), logging.Any("array", a))

	parent.Report("This log entry will go to the BI pipeline and be loaded into our warehouse if kinesis is enabled!")
}
