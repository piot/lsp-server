package lspserv

import (
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/sourcegraph/jsonrpc2"
)

func listen(addr string) (*net.Listener, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to bind to address %s: %v", addr, err)
		return nil, err
	}

	if os.Getenv("TLS_CERT") != "" && os.Getenv("TLS_KEY") != "" {
		cert, err := tls.X509KeyPair([]byte(os.Getenv("TLS_CERT")), []byte(os.Getenv("TLS_KEY")))
		if err != nil {
			return nil, err
		}
		listener = tls.NewListener(listener, &tls.Config{
			Certificates: []tls.Certificate{cert},
		})
	}
	return &listener, nil
}

type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}

func newHandler(implementationHandler Handler) (jsonrpc2.Handler, io.Closer) {
	return NewLspRequests(implementationHandler), ioutil.NopCloser(strings.NewReader(""))
}

func acceptConnections(addr string, implementationHandler Handler) error {
	{
		connectionCount := 0
		lis, err := listen(addr)
		if err != nil {
			return err
		}

		defer (*lis).Close()

		for {
			conn, err := (*lis).Accept()
			if err != nil {
				return err
			}
			connectionCount = connectionCount + 1
			connectionID := connectionCount
			log.Printf("lsp-server: received incoming connection #%d\n", connectionID)
			handler, closer := newHandler(implementationHandler)
			var connOpt []jsonrpc2.ConnOpt

			jsonrpc2Connection := jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(conn, jsonrpc2.VSCodeObjectCodec{}), handler, connOpt...)
			go func() {
				<-jsonrpc2Connection.DisconnectNotify()
				err := closer.Close()
				if err != nil {
					log.Println(err)
				}
				log.Printf("lsp-server: connection #%d closed\n", connectionID)
			}()
		}

	}
}

func RunForever(addr string, implementationHandler Handler) error {

	log.Println("lsp-server: listening on", addr)

	//cfg := langserver.Config{}

	go acceptConnections(addr, implementationHandler)

	mode := "stdin"

	switch mode {
	case "stdin":
		{
			var connOpt []jsonrpc2.ConnOpt
			handler, closer := newHandler(implementationHandler)
			<-jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{}), handler, connOpt...).DisconnectNotify()
			err := closer.Close()
			if err != nil {
				log.Println(err)
			}
		}
	case "jsonrpc":

	}

	return nil

}
