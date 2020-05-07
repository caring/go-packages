package main

import "github.com/caring/go-packages/pkg/logging"

func main() {
	t := true
	parent, err := logging.NewLogger(&logging.Config{
		LoggerName:          "logger-1",
		ServiceName:         "call-scoring",
		LogLevel:            "DEBUG",
		EnableDevLogging:    &t,
		KinesisStreamName:   "stream-1",
		KinesisPartitionKey: "some-key",
		DisableKinesis:      &t,
	})

	if err != nil {
		panic(err)
	}

	child := parent.NewChild(
		&logging.InternalFields{
			Endpoint:     "Some-endpoint",
			IsReportable: &t,
		},
		logging.NewStringField("child", "this wont be in the parent"),
	)

	child.Warn("sample message", logging.NewInt64Field("fieldA", 3))

	child.AppendAdditionalFields(logging.NewBoolField("fieldB", true), logging.NewStringField("fieldB", "helloworld"))

	parent.Warn("here's another waring, see no child field!")

	child.SetIsReportable(true)
	child.SetServiceName("some service")

	child.Warn("I'm the child, final warning... see my fields have changed!", logging.NewFloat64sField("floats", []float64{1.0, 2.0, 3.0}))

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

	parent.Warn("Look how any field can marshal complex objects, this is expensive!", logging.NewAnyField("map", m), logging.NewAnyField("array", a))
}
