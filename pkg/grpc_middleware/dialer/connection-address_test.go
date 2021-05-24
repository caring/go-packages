package dialer

import (
	"testing"

	"github.com/matryer/is"
)

func TestReadConnectionAddress(t *testing.T) {
	is := is.New(t)

	// Empty config should return nil and error
	cfg, err := ReadConnectionAddress("")
	is.True(err != nil)
	is.Equal(cfg, nil)
	is.Equal(cfg.String(), "<nil>")
	err = cfg.ApplyToBuilder(&Builder{})
	is.True(err != nil)

	// Populated config should parse
	cfg, err = ReadConnectionAddress("tcp://localhost:1234")
	is.NoErr(err)
	is.Equal(cfg.String(), "tcp://localhost:1234")
	tls, err := cfg.loadTLS(nil)
	is.True(err != nil)
	is.Equal(tls, nil)

	// Populated config should apply to builder
	err = cfg.ApplyToBuilder(&Builder{})
	is.NoErr(err)

	// Applying to nil should fail
	err = cfg.ApplyToBuilder(nil)
	is.True(err != nil)

	// Applying to nil interface should fail
	err = cfg.ApplyToBuilder((*Builder)(nil))
	is.True(err != nil)

	// tls string should parse
	cfg, err = ReadConnectionAddress("tls://localhost:1234")
	is.NoErr(err)
	is.Equal(cfg.String(), "tls://localhost:1234")
	tls, err = cfg.loadTLS(nil)
	is.NoErr(err)
	is.Equal(tls.InsecureSkipVerify, false)

	// tls string with skip_verify should parse
	cfg, err = ReadConnectionAddress("tls://localhost:1234?skip_verify=true")
	is.NoErr(err)
	is.Equal(cfg.String(), "tls://localhost:1234?skip_verify=true")
	tls, err = cfg.loadTLS(nil)
	is.NoErr(err)
	is.Equal(tls.InsecureSkipVerify, true)
}
