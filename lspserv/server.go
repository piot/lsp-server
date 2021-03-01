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

type StdInOutReadWriteCloser struct {
	logOutput bool
}

func (s StdInOutReadWriteCloser) Read(p []byte) (int, error) {
	n, err := os.Stdin.Read(p)

	if s.logOutput {
		log.Printf("read: '%v'\n", string(p[0:n]))
	}

	return n, err
}

func (s StdInOutReadWriteCloser) Write(p []byte) (int, error) {
	if s.logOutput {
		log.Printf("write: '%v'\n", string(p))
	}

	return os.Stdout.Write(p)
}

func (s StdInOutReadWriteCloser) Close() error {
	if s.logOutput {
		log.Printf("close\n")
	}

	if err := os.Stdin.Close(); err != nil {
		return err
	}

	return os.Stdout.Close()
}

type Service interface {
	RunUntilClose(rwc io.ReadWriteCloser, logOutput bool) error
}

type serviceWrapper struct {
	lspRequests *HandleLspRequests
}

func NewService(implementationHandler Handler) Service {
	return &serviceWrapper{lspRequests: NewLspRequests(implementationHandler)}
}

func (s *serviceWrapper) RunUntilClose(rwc io.ReadWriteCloser, logOutput bool) error {
	var connOpt []jsonrpc2.ConnOpt

	stdErrLogger := log.New(os.Stderr, "", log.LstdFlags)
	if logOutput {
		connOpt = append(connOpt, jsonrpc2.LogMessages(stdErrLogger))
	}

	closer := ioutil.NopCloser(strings.NewReader(""))

	connection := jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(rwc,
		jsonrpc2.VSCodeObjectCodec{}), s.lspRequests, connOpt...)

	<-connection.DisconnectNotify()

	err := closer.Close()
	if err != nil {
		log.Println(err)
	}

	return nil
}
