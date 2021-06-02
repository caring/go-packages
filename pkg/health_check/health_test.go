package health_check

import "testing"
import "github.com/caring/go-packages/v2/pkg/logging"
import "github.com/stretchr/testify/assert"

func Test_NewEndpoint(t *testing.T) {
	service = "emailbroker"
	branch = "master"
	sha1 = "d6fa73ae5348b232c785ca1596e1cd3d52be115c"
	tag = "v1.0.3"
	goVersion = "go1.13.8"
	l := logging.Logger{}

	endPoint := NewEndpoint(&l)

	assert.Equal(t, endPoint.Service, service)
	assert.Equal(t, endPoint.Branch, branch)
	assert.Equal(t, endPoint.SHA1, sha1)
	assert.Equal(t, endPoint.Tag, tag)
	assert.Equal(t, endPoint.GoVersion, goVersion)
}