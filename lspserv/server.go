package lspserv

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/piot/jsonrpc2"
)

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

func RunForever(addr string, implementationHandler Handler) error {
	var connOpt []jsonrpc2.ConnOpt

	stdErrLogger := log.New(os.Stderr, "", log.LstdFlags)
	connOpt = append(connOpt, jsonrpc2.LogMessages(stdErrLogger))

	handler, closer := newHandler(implementationHandler)

	connection := jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(stdrwc{},
		jsonrpc2.VSCodeObjectCodec{}), handler, connOpt...)

	<-connection.DisconnectNotify()

	err := closer.Close()
	if err != nil {
		log.Println(err)
	}

	return nil
}
