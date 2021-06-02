package health_check

import (
	"runtime"
	"testing"

	"github.com/caring/go-packages/v2/pkg/logging"
	"github.com/stretchr/testify/assert"
)


func Test_NewEndpoint(t *testing.T) {
	Service = "emailbroker"
	Branch = "master"
	SHA1 = "d6fa73ae5348b232c785ca1596e1cd3d52be115c"
	Tag = "v1.0.3"
	goVersion := runtime.Version()
	l := logging.Logger{}

	endPoint := NewEndpoint(&l)
	assert.Equal(t, endPoint.Service, Service)
	assert.Equal(t, endPoint.Branch, Branch)
	assert.Equal(t, endPoint.SHA1, SHA1)
	assert.Equal(t, endPoint.Tag, Tag)
	assert.Equal(t, endPoint.GoVersion, goVersion)

	Tag = ""

	endPoint = NewEndpoint(&l)
	assert.Equal(t, endPoint.Service, Service)
	assert.Equal(t, endPoint.Branch, Branch)
	assert.Equal(t, endPoint.SHA1, SHA1)
	assert.Equal(t, endPoint.Tag, "N/A")
	assert.Equal(t, endPoint.GoVersion, goVersion)
}
