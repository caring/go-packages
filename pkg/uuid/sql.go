package uuid

import (
	"database/sql/driver"
	_ "fmt"
	"github.com/caring/go-packages/v2/pkg/errors"
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

func (uuid *UUID) AssignTo(dst interface{}) error {

	// switch v := dst.(type) {
	// case *[16]byte:
	// 	*v = src.Bytes
	// 	return nil
	// case *[]byte:
	// 	*v = make([]byte, 16)
	// 	copy(*v, src.Bytes[:])
	// 	return nil
	// case *string:
	// 	*v = Parse(string(src.Bytes))
	// 	return nil
	// default:
	// 	if nextDst, retry := GetAssignToDstType(v); retry {
	// 		return src.AssignTo(nextDst)
	// 	}
	// }

	return errors.Errorf("cannot assign %v into %T", uuid, dst)
}

// Set converts and assigns src to itself. Value takes ownership of src.
func (uuid *UUID) Set(src interface{}) error {
	return nil
}

// Get returns the simplest representation of Value. Get may return a pointer to an internal value but it must never
// mutate that value. e.g. If Get returns a []byte Value must never change the contents of the []byte.
func (uuid *UUID) Get() interface{} {
	return uuid
}
