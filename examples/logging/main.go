package main

import "github.com/caring/go-packages/pkg/logging"

func main() {
	l, err := logging.InitLogging(false, "logger-1", "stream-1", "call-scoring", "aws-key", "aws-secret", "aws-region", true, true)

	if err != nil {
		panic(err)
	}

	l.NewChild("", "endpoint2", 3, 3, 2, 1, false, nil)

	l.Warn("sample message", map[string]logging.Field{"fieldA": logging.NewInt64Field("fieldA", 3)})

	l.AppendAdditionalData(map[string]logging.Field{"fieldB": logging.NewBoolField("fieldB", true)})

	l.Warn("here's another waring", nil)

	l.SetIsReportable(true)
	l.SetAdditionalData(nil)

	l.Warn("final warning", nil)
}