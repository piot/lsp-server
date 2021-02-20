package main

import (
	"context"

	"github.com/piot/jsonrpc2"
	"github.com/piot/lsp-server/lspserv"
	"github.com/sourcegraph/go-lsp"
)

type MyHandler struct {
}

func (m *MyHandler) HandleHover(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (*lsp.Hover, error) {
	return &lsp.Hover{
		Contents: []lsp.MarkedString{
			lsp.RawMarkedString("This is just a hello from lsp server"),
		},
		Range: &lsp.Range{
			Start: lsp.Position{
				Line:      0,
				Character: 0,
			},
			End: lsp.Position{
				Line:      0,
				Character: 0,
			},
		},
	}, nil
}

func (m *MyHandler) ShutDown() {

}

func (m *MyHandler) ResetCaches(lock bool) {

}

func (m *MyHandler) Reset() error {
	return nil
}

func main() {
	lspserv.RunForever(":6009", &MyHandler{})
}
