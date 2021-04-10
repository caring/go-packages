package graphid

import (
	"encoding/base64"
	"github.com/caring/go-packages/v2/pkg/errors"
	"strconv"
	"strings"
	// "github.com/caring/go-packages/pkg/errors"
)

// EncodePseudoGuid creates a pseudo GUID for integer based IDs, which are only used internally
// within a single service and not referenced by external services. This practice is recommended
// by GraphQL spec, and they recommend to Base64 encode the model type string + the ID
//
// ex mypackage.MyModel{ID: id} -> EncodePseudoGuid("MyModel", id)
func EncodePseudoGuid(typeString string, id int64) string {
	s := strconv.Itoa(int(id))
	return base64.StdEncoding.EncodeToString([]byte(typeString + s))
}

// DecodePseudoGuid decodes a pseudo GUID created by EncodePseudoGuid
func DecodePseudoGuid(typeString string, pseudo string) (int64, error) {
	decoded, err := base64.StdEncoding.DecodeString(pseudo)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	s := strings.Replace(string(decoded), typeString, "", -1)
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	return i, nil
}
