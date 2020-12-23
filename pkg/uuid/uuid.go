package uuid

import (
	"database/sql"
	"github.com/caring/go-packages/pkg/errors"
	goouid "github.com/google/uuid"
)

type UUID struct {
	goouid.UUID
}

func New() UUID {
	uuid := goouid.New()
	return UUID{UUID: uuid}
}

func NewRandom() (UUID, error) {
	uuid, err := goouid.NewRandom()
	if err != nil {
		return UUID{}, err
	}
	return UUID{UUID: uuid}, nil
}

func fromGoogleUuid(uuid goouid.UUID) UUID {
	return UUID{UUID: uuid}
}

// Must returns uuid if err is nil and panics otherwise.
func Must(uuid UUID, err error) UUID {
	return fromGoogleUuid(goouid.Must(uuid.UUID, err))
}

func MustParse(s string) UUID {
	uid := goouid.MustParse(s)
	return fromGoogleUuid(uid)
}

// ParseUUID parses a UUID string to a byte slice UUID. if the string is empty it return uuid.Nil
func Parse(s string) (UUID, error) {
	if s == "" {
		return UUID{}, nil
	}
	uid, err := goouid.Parse(s)
	if err != nil {
		return UUID{}, errors.WithStack(err)
	}
	return fromGoogleUuid(uid), nil
}

// ParseUUIDs is a convenience method to parse multiple strings to UUIDs,
// if any error is returned then a nil slice is returned. Each empty string is parsed to
// a 0 value UUID
func ParseUUIDs(ss []string) ([]UUID, error) {
	parsed := make([]UUID, len(ss))
	for i, s := range ss {
		uid, err := Parse(s)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		parsed[i] = uid
	}
	return parsed, nil
}

func ParseBytes(b []byte) (UUID, error) {
	uid, err := goouid.ParseBytes(b)
	if err != nil {
		return UUID{}, errors.WithStack(err)
	}
	return fromGoogleUuid(uid), nil
}

func (uuid UUID) IsNil() bool {
	return uuid.UUID.ID() == 0
}

func (uuid UUID) String() string {
	return uuid.UUID.String()
}

func (uuid UUID) NullString() sql.NullString {
	if uuid.IsNil() {
		return sql.NullString{String: uuid.String(), Valid: false}
	}
	return sql.NullString{String: uuid.String(), Valid: true}
}

func (uuid UUID) URN() string {
	return string(uuid.UUID.URN())
}

func (uuid UUID) Variant() goouid.Variant {
	return uuid.UUID.Variant()
}

// Version returns the version of uuid.
func (uuid UUID) Version() goouid.Version {
	return uuid.UUID.Version()
}
