package sentry

import (
	"log"
	"os"
)
import sg "github.com/getsentry/sentry-go"

func init() {
	sentryDsn := os.Getenv("SENTRY_DSN")
	env := os.Getenv("ENV")
	if env != "development" {
		err := sg.Init(sentry.ClientOptions{
			Dsn:         sentryDsn,
			Environment: env,
		})
		if err != nil {
			sg.CaptureException(err)
			log.Fatalf("sentry.Init: %s", err)
		}
		defer sg.Flush(2 * time.Second)
	}
}

func ErrorAsException(err error) {
	
}
