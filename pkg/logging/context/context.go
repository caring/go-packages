package ctx

import (
	"context"
	"errors"

	"github.com/caring/go-packages/pkg/logging"
)

// Extract gets a Logger instance from a context.Context
func Extract(ctx context.Context) (*logging.Logger, error) {
	return logging.NewNopLogger(), errors.New("Not implemented")
}
