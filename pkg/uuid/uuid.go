package uuid

import (
	"github.com/caring/go-packages/pkg/errors"
	goouid "github.com/google/uuid"
)

type UUID [16]byte

// ParseUUID parses a UUID string to a byte slice UUID. if the string is empty it return uuid.Nil
func ParseUUID(s string) (UUID, error) {
	if s == "" {
		return goouid.Nil, nil
	}
	uid, err := goouid.Parse(s)
	if err != nil {
		return goouid.Nil, errors.WithStack(err)
	}
	return uid, nil
}

// ParseUUIDs is a convenience method to parse multiple strings to UUIDs,
// if any error is returned then a nil slice is returned. Each empty string is parsed to
// a 0 value UUID
func ParseUUIDs(ss []string) ([]UUID, error) {
	parsed := make([]UUID, len(ss))
	for i, s := range ss {
		uid, err := ParseUUID(s)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		parsed[i] = uid
	}
	return parsed, nil
}
