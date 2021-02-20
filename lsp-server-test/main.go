package main

import (
	"github.com/piot/go-lsp"

	"github.com/piot/lsp-server/lspserv"
)




type MyHandler struct {
}

func (m *MyHandler) HandleHover(params lsp.TextDocumentPositionParams, conn lspserv.Connection) (*lsp.Hover, error) {
	return &lsp.Hover{
		Contents: lsp.MarkupContent{
			Kind: lsp.MUKMarkdown,
			Value: "this is **markup** content\n---\nIs this the last line?",
		},
		Range: &lsp.Range{
			Start: lsp.Position{
				Line:      0,
				Character: 0,
			},
			End: lsp.Position{
				Line:      0,
				Character: 24,
			},
		},
	}, nil
}

func (m *MyHandler) HandleGotoDefinition(params lsp.TextDocumentPositionParams, conn lspserv.Connection) (*lsp.Location, error) {
	return &lsp.Location{
		URI: params.TextDocument.URI,
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      0,
				Character: 0,
			},
			End: lsp.Position{
				Line:      0,
				Character: 40,
			},
		},
	}, nil
}

// Called after GotoDefinition is used?
func (m *MyHandler) HandleTextDocumentReferences(params lsp.ReferenceParams, conn lspserv.Connection) ([]*lsp.Location, error) {
	return []*lsp.Location{}, nil
}

func (m *MyHandler) HandleTextDocumentSymbol(params lsp.DocumentSymbolParams,conn lspserv.Connection) ([]*lsp.DocumentSymbol, error) {
	diagnosticParams := lsp.PublishDiagnosticsParams{
		URI:         params.TextDocument.URI,
		Diagnostics: []lsp.Diagnostic{
			{
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      2,
						Character: 0,
					},
					End: lsp.Position{
						Line:      2,
						Character: 5,
					},
				},
				Severity: lsp.Warning,
				Code:     "A1233",
				Source:   "swamp",
				Message:  "You can not provide this crappy code",
			},
		},
	}

	conn.PublishDiagnostics(diagnosticParams)

	return []*lsp.DocumentSymbol{
		{
			Name:   "name",
			Detail: "String name",
			Kind:   lsp.SKProperty,
			Tags:   nil,
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      0,
					Character: 0,
				},
				End: lsp.Position{
					Line:      0,
					Character: 4,
				},
			},
			SelectionRange: lsp.Range{
				Start: lsp.Position{
					Line:      0,
					Character: 0,
				},
				End: lsp.Position{
					Line:      0,
					Character: 4,
				},
			},
			Children: nil,
		},
		{
			Name:   "2",
			Detail: "Int",
			Kind:   lsp.SKNumber,
			Tags:   nil,
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      0,
					Character: 6,
				},
				End: lsp.Position{
					Line:      0,
					Character: 6,
				},
			},
			SelectionRange: lsp.Range{
				Start: lsp.Position{
					Line:      0,
					Character: 6,
				},
				End: lsp.Position{
					Line:      0,
					Character: 6,
				},
			},
			Children: nil,
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
