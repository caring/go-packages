// +build !go1.16
//go:build !go1.16

package dialer

import (
	"crypto/tls"
	"io/ioutil"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

const (
	fsCompat = false
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
}

func (b *Builder) WithFS(fs interface{}) {
	panic("not implemented pre go 1.16")
}

func (b *Builder) GetFS() interface{} {
	return nil
}

func readFile(b *Builder, filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(f)
}

func dirFS(path string) interface{} {
	return nil
}
