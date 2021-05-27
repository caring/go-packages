// +build go1.16
//go:build go1.16

package dialer

import (
	"crypto/tls"
	"io"
	"io/fs"
	"io/ioutil"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

const (
	fsCompat = true
)

type Builder struct {
	options         []grpc.DialOption
	enabledBlocking bool
	connectParams   grpc.ConnectParams
	keepAliveParams keepalive.ClientParameters
	credentials     credentials.PerRPCCredentials
	uinterceptors   []grpc.UnaryClientInterceptor
	sinterceptors   []grpc.StreamClientInterceptor
	tlsConfig       *tls.Config
	dns             *string
	port            *uint16
	fs              fs.FS
}

// WithFS will set the filesystem to use when loading resouces. If not set will fallback to using os.Open
// pre go 1.16 does not support alternate filesystems only os.Open is used.
func (b *Builder) WithFS(fs fs.FS) {
	b.fs = fs
}

func (b *Builder) GetFS() fs.FS {
	return b.fs
}

func readFile(b *Builder, filename string) ([]byte, error) {
	var f io.Reader
	var err error

	fsys := b.GetFS()
	if fsys != nil {
		f, err = fsys.Open(filename)
	} else {
		f, err = os.Open(filename)
	}
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(f)
}

func dirFS(path string) fs.FS {
	return os.DirFS(path)
}