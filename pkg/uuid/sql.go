package uuid

import (
	"database/sql/driver"
	"github.com/caring/go-packages/pkg/errors"
	_ "github.com/google/uuid"
)

func (uuid *UUID) Scan(src interface{}) error {
	if err := uuid.UUID.Scan(src); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (uuid UUID) Value() (driver.Value, error) {
	return uuid.String(), nil
}
