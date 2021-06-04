package dialer

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestLoadCerts(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		IsCA:                  true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	certOut, err := ioutil.TempFile("", "cert.pem")
	if err != nil {
		t.Fatalf("Failed to open cert.pem for writing: %v", err)
	}
	defer os.Remove(certOut.Name())

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		t.Fatalf("Failed to write data to cert.pem: %v", err)
	}
	if err := certOut.Close(); err != nil {
		t.Fatalf("Error closing cert.pem: %v", err)
	}
	t.Logf("wrote %s\n", certOut.Name())

	is := is.New(t)

	// read CA from filesystem
	cfg, err := ReadConnectionAddress("tls://localhost:1234?ca_file=" + certOut.Name())
	is.NoErr(err)
	is.Equal(cfg.String(), "tls://localhost:1234?ca_file="+strings.Replace(certOut.Name(), "/", "%2F", -1))
	tlsConfig, err := cfg.loadTLS(&Builder{})
	is.NoErr(err)
	is.Equal(len(tlsConfig.RootCAs.Subjects()), 1)
	is.Equal(tlsConfig.RootCAs.Subjects()[0], []byte{48, 18, 49, 16, 48, 14, 6, 3, 85, 4, 10, 19, 7, 65, 99, 109, 101, 32, 67, 111}) // "Acme Co"

	// read CA from fs.FS
	if fsCompat {
		fname := filepath.Base(certOut.Name())

		cfg, err := ReadConnectionAddress("tls://localhost:1234?ca_file=" + fname)
		is.NoErr(err)
		is.Equal(cfg.String(), "tls://localhost:1234?ca_file="+strings.Replace(fname, "/", "%2F", -1))
		b := &Builder{}
		b.WithFS(dirFS(filepath.Dir(certOut.Name())))
		tlsConfig, err := cfg.loadTLS(b)
		is.NoErr(err)
		is.Equal(len(tlsConfig.RootCAs.Subjects()), 1)
		is.Equal(tlsConfig.RootCAs.Subjects()[0], []byte{48, 18, 49, 16, 48, 14, 6, 3, 85, 4, 10, 19, 7, 65, 99, 109, 101, 32, 67, 111}) // "Acme Co"
	}
}
