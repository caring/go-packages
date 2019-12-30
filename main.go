package main

import "github.com/caring/go-packages/pkg/logging"


func main() {
	l, err := logging.InitLogging(false, "testing", "testing", "call-scoring", "AKIAYDFPRUWTSXGVTH5B", "LBEensrEKur30Yl8x0WwblACK2qZ/fek49BD+iNc", "us-east-1", true, true)

	if err != nil {
		panic(err)
	}

	l.Warn("sample warning message", nil)

	l.SetAdditionalData(map[string]logging.Field{
		"sampleContentA": logging.NewStringField("sampleContentA", "blah"),
	})

	l.Warn("here's another warning", nil)

	l.Warn("more warning", map[string]logging.Field{
		"fieldA": logging.NewInt64Field("intField", 3),
	})
}