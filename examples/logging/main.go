package main

import "github.com/caring/go-packages/pkg/logging"

func main() {
	parent, err := logging.InitLogging(false, "logger-1", "stream-1", "call-scoring", "aws-key", "aws-secret", "aws-region", true, false)

	if err != nil {
		panic(err)
	}

	child := parent.NewChild("", "endpoint2", 3, 3, 2, 1, false, []logging.Field{logging.NewStringField("child", "this wont be in the parent")})

	child.Warn("sample message", []logging.Field{logging.NewInt64Field("fieldA", 3)}...)

	child.AppendAdditionalData([]logging.Field{logging.NewBoolField("fieldB", true)})

	parent.Warn("here's another waring, see no child field!")

	child.SetIsReportable(true)
	child.SetServiceID("some service")

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
