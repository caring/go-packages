package main

import "github.com/caring/go-packages/pkg/logging"

func main() {
	l, err := logging.InitLogging(false, "logger-1", "stream-1", "call-scoring", "aws-key", "aws-secret", "aws-region", true, false)

	if err != nil {
		panic(err)
	}

	l.NewChild("", "endpoint2", 3, 3, 2, 1, false, nil)

	l.Warn("sample message", []logging.Field{logging.NewInt64Field("fieldA", 3)}...)

	l.AppendAdditionalData([]logging.Field{logging.NewBoolField("fieldB", true)})

	l.Warn("here's another waring")

	l.SetIsReportable(true)
	l.SetAdditionalData(nil)

	l.Warn("final warning")
}
