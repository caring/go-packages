package exit

import (
	"testing"

	"gotest.tools/v3/assert"
)

func Test_Stub(t *testing.T) {
	tests := []struct {
		f    func()
		want bool
	}{
		{Exit, true},
		{func() {}, false},
	}

	for _, tt := range tests {
		s := WithStub(tt.f)
		assert.Equal(t, tt.want, s.Exited, "Stub captured unexpected exit value.")
	}
}
