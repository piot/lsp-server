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
			Kind:  lsp.MUKMarkdown,
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

func (m *MyHandler) HandleGotoTypeDefinition(params lsp.TextDocumentPositionParams, conn lspserv.Connection) (*lsp.Location, error) {
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

func (m *MyHandler) HandleTextDocumentCompletion(params lsp.CompletionParams, conn lspserv.Connection) (*lsp.CompletionList, error) {
	return &lsp.CompletionList{
		IsIncomplete: false,
		Items: []lsp.CompletionItem{
			{
				Label: "filterMap",
				Kind:  lsp.CIKFunction,
				// Tags:
				Detail:        "filterMap filters out and maps a lot of stuff.",
				Documentation: "This is a doc comment",
				//Preselect
				SortText:         "",
				FilterText:       "",
				InsertText:       "",
				InsertTextFormat: lsp.ITFPlainText,
				// InsertTextMode
				TextEdit: nil,
				// AdditionalTextEdit
				// CommitCharacters
				// Command:
				Data: nil,
			},
		},
	}, nil
}

func (m *MyHandler) HandleTextDocumentSignatureHelp(params lsp.TextDocumentPositionParams, conn lspserv.Connection) (*lsp.SignatureHelp, error) {
	return &lsp.SignatureHelp{
		Signatures: []lsp.SignatureInformation{{
			Label:         "(Int -> a -> b) -> List a -> List b",
			Documentation: "indexes a lot of maps",
			Parameters: []lsp.ParameterInformation{
				{
					Label:         "(Int -> a -> b)",
					Documentation: "a function that takes an int and converts from a to b",
				},
				{
					Label:         "List a",
					Documentation: "a list of as",
				},
			},
		},
		},
		ActiveSignature: 0,
		ActiveParameter: 0,
	}, nil
}

func (m *MyHandler) HandleTextDocumentSymbol(params lsp.DocumentSymbolParams, conn lspserv.Connection) ([]*lsp.DocumentSymbol, error) {
	diagnosticParams := lsp.PublishDiagnosticsParams{
		URI: params.TextDocument.URI,
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

func (m *MyHandler) HandleCodeAction(params lsp.CodeActionParams, conn lspserv.Connection) (*lsp.CodeAction, error) {
	return &lsp.CodeAction{
		Title: "",
		Kind:  "",
		Diagnostics: []lsp.Diagnostic{
			{
				Range:              lsp.Range{},
				Severity:           0,
				Code:               "",
				CodeDescription:    &lsp.CodeDescription{},
				Source:             "",
				Message:            "",
				Tags:               []lsp.DiagnosticTag{},
				RelatedInformation: []lsp.DiagnosticRelatedInformation{},
				Data:               nil,
			},
		},
	}, nil
}

func (m *MyHandler) HandleCodeActionResolve(params lsp.CodeAction, conn lspserv.Connection) (*lsp.CodeAction, error) {
	return &params, nil
}

func (m *MyHandler) HandleRename(params lsp.RenameParams) (*lsp.WorkspaceEdit, error) {
	return &lsp.WorkspaceEdit{
		Changes: map[string][TextEdit]{
			
		}
	}
}

func main() {
	lspserv.RunForever(":6009", &MyHandler{})
}
