package main

import (
	"fmt"

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
func (m *MyHandler) HandleReferences(params lsp.ReferenceParams, conn lspserv.Connection) ([]*lsp.Location, error) {
	return []*lsp.Location{}, nil
}

func (m *MyHandler) HandleCompletion(params lsp.CompletionParams, conn lspserv.Connection) (*lsp.CompletionList, error) {
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

func (m *MyHandler) HandleDidOpen(params lsp.DidOpenTextDocumentParams, conn lspserv.Connection) error {
	return nil
}

func (m *MyHandler) HandleLinkedEditingRange(params lsp.LinkedEditingRangeParams, conn lspserv.Connection) (*lsp.LinkedEditingRanges, error) {
	return nil, nil
}

func (m *MyHandler) HandleCompletionItemResolve(params lsp.CompletionItem, conn lspserv.Connection) (*lsp.CompletionItem, error) {
	return &params, nil
}

func (m *MyHandler) HandleFindReferences(params lsp.ReferenceParams, conn lspserv.Connection) ([]*lsp.Location, error) {
	return []*lsp.Location{

		{
			URI: params.TextDocument.URI,
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      3,
					Character: 0,
				},
				End: lsp.Position{
					Line:      3,
					Character: 5,
				},
			},
		},
		{
			URI: params.TextDocument.URI,
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      5,
					Character: 0,
				},
				End: lsp.Position{
					Line:      5,
					Character: 5,
				},
			},
		},
	}, nil
}

func (m *MyHandler) HandleFormatting(params lsp.DocumentFormattingParams, conn lspserv.Connection) ([]*lsp.TextEdit, error) {
	return []*lsp.TextEdit{
		{
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      0,
					Character: 0,
				},
				End: lsp.Position{
					Line:      0,
					Character: 0,
				},
			},
			NewText: "This is now formatted. Happy?",
		},
	}, nil
}

func (m *MyHandler) HandleGotoDeclaration(params lsp.DeclarationOptions, conn lspserv.Connection) (*lsp.Location, error) {
	return nil, fmt.Errorf("this language doesn't support goto declaration")
}

func (m *MyHandler) HandleGotoImplementation(params lsp.TextDocumentPositionParams, conn lspserv.Connection) (*lsp.Location, error) {
	return nil, fmt.Errorf("this language doesn't support goto implementation")
}

func (m *MyHandler) HandleHighlights(params lsp.DocumentHighlightParams, conn lspserv.Connection) ([]*lsp.DocumentHighlight, error) {
	return []*lsp.DocumentHighlight{
		{
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      0,
					Character: 0,
				},
				End: lsp.Position{
					Line:      2,
					Character: 3,
				},
			},
			Kind: lsp.Text,
		},
	}, nil
}

func (m *MyHandler) HandleSignatureHelp(params lsp.TextDocumentPositionParams, conn lspserv.Connection) (*lsp.SignatureHelp, error) {
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

func (m *MyHandler) HandleDidChangeWatchedFiles(params lsp.DidChangeWatchedFilesParams, conn lspserv.Connection) error {
	return nil
}

func (m *MyHandler) HandleSymbol(params lsp.DocumentSymbolParams, conn lspserv.Connection) ([]*lsp.DocumentSymbol, error) {
	/*
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

	*/

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
		Changes: map[string][]lsp.TextEdit{},
	}, nil
}

func (m *MyHandler) HandleCodeLens(params lsp.CodeLensParams, conn lspserv.Connection) ([]*lsp.CodeLens, error) {
	return []*lsp.CodeLens{{
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      4,
				Character: 0,
			},
			End: lsp.Position{
				Line:      4,
				Character: 4,
			},
		},
		Command: lsp.Command{
			Title:     "Some Command here",
			Command:   "swamp.somecommand",
			Arguments: nil,
		},
		Data: nil,
	},
	}, nil
}

func (m *MyHandler) HandleSemanticTokensFull(params lsp.SemanticTokensParams, conn lspserv.Connection) (*lsp.SemanticTokens, error) {
	//						TokenTypes:     []string{"type", "enum", "struct", "typeParameter", "parameter"},
	//TokenModifiers: []string{"declaration", "definition"},
	// deltaLine, deltaColumn, Length, tokenType, tokenModifierFlags
	return &lsp.SemanticTokens{
		ResultId: "",
		Data: []uint{
			0, 0, 5, 0, 1,
			2, 0, 5, 3, 2,
		},
	}, nil
}

func (m *MyHandler) HandleCodeLensResolve(params lsp.CodeLens, conn lspserv.Connection) (*lsp.CodeLens, error) {
	return &lsp.CodeLens{
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      4,
				Character: 0,
			},
			End: lsp.Position{
				Line:      4,
				Character: 4,
			},
		},
		Command: lsp.Command{
			Title:     "Some Command here",
			Command:   "swamp.somecommand",
			Arguments: nil,
		},
		Data: nil,
	}, nil
}

func main() {
	testHandler := &MyHandler{}
	service := lspserv.NewService(testHandler)

	service.RunUntilClose(lspserv.StdInOutReadWriteCloser{}, true)
}
