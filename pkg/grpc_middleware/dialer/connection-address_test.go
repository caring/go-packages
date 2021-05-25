package dialer

import (
	"testing"

	"github.com/matryer/is"
)

func TestReadConnectionAddress(t *testing.T) {
	is := is.New(t)

	// Check empty address returns nil
	cfg, err := ReadConnectionAddress("")
	is.True(err != nil)
	is.True(cfg.String() == "<nil>")

	// check dns: string
	cfg, err = ReadConnectionAddress("dns:example.dev.caring.com")
	is.NoErr(err)
	is.Equal(cfg.String(), "tls://example.dev.caring.com:443")

	// check tls string
	cfg, err = ReadConnectionAddress("tls://example.dev.caring.com")
	is.NoErr(err)
	is.Equal(cfg.String(), "tls://example.dev.caring.com:443")

	// check http scheme to tcp
	cfg, err = ReadConnectionAddress("http://example.dev.caring.com")
	is.NoErr(err)
	is.Equal(cfg.String(), "tcp://example.dev.caring.com:80")

	// check no scheme defaults to tls
	cfg, err = ReadConnectionAddress("example.dev.caring.com")
	is.NoErr(err)
	is.Equal(cfg.String(), "tls://example.dev.caring.com:443")

	// check https scheme to tls
	cfg, err = ReadConnectionAddress("https://example.dev.caring.com")
	is.NoErr(err)
	is.Equal(cfg.String(), "tls://example.dev.caring.com:443")
}

func TestApplyBuilder(t *testing.T) {
	is := is.New(t)

	// check https scheme to tls
	cfg, err := ReadConnectionAddress("https://example.dev.caring.com")
	is.NoErr(err)
	is.Equal(cfg.String(), "tls://example.dev.caring.com:443")

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
	tlsConfig, err := cfg.loadTLS(nil)
	is.NoErr(err)
	is.Equal(tlsConfig.InsecureSkipVerify, false)

	// tls string with skip_verify should parse
	cfg, err = ReadConnectionAddress("tls://localhost:1234?skip_verify=true")
	is.NoErr(err)
	is.Equal(cfg.String(), "tls://localhost:1234?skip_verify=true")
	tlsConfig, err = cfg.loadTLS(nil)
	is.NoErr(err)
	is.Equal(tlsConfig.InsecureSkipVerify, true)
}
