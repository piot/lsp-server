package main

import (
	"context"

	"github.com/piot/lsp-server/lspserv"
	"github.com/sourcegraph/go-langserver/langserver"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
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

func (m *MyHandler) Reset(params *langserver.InitializeParams) error {
	return nil
}

func main() {
	lspserv.RunForever(":6009", &MyHandler{})
}
