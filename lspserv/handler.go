package lspserv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/piot/go-lsp"
	"github.com/piot/jsonrpc2"
)

type Connection interface {
	PublishDiagnostics(params lsp.PublishDiagnosticsParams) error
	//RequestCodeLensRefresh() error
}

type Handler interface {
	Reset() error
	ResetCaches(lock bool)
	ShutDown()
	HandleHover(params lsp.TextDocumentPositionParams, conn Connection) (*lsp.Hover, error)
	HandleGotoDefinition(params lsp.TextDocumentPositionParams, conn Connection) (*lsp.Location, error)
	HandleGotoDeclaration(params lsp.DeclarationOptions, conn Connection) (*lsp.Location, error)
	HandleGotoTypeDefinition(params lsp.TextDocumentPositionParams, conn Connection) (*lsp.Location, error)
	HandleGotoImplementation(params lsp.TextDocumentPositionParams, conn Connection) (*lsp.Location, error)
	HandleFindReferences(params lsp.ReferenceParams, conn Connection) ([]*lsp.Location, error)
	HandleSymbol(params lsp.DocumentSymbolParams, conn Connection) ([]*lsp.DocumentSymbol, error) // Used for outline
	HandleLinkedEditingRange(params lsp.LinkedEditingRangeParams, conn Connection) (*lsp.LinkedEditingRanges, error)
	HandleCompletion(params lsp.CompletionParams, conn Connection) (*lsp.CompletionList, error) // Intellisense when pressing '.'.
	HandleCompletionItemResolve(params lsp.CompletionItem, conn Connection) (*lsp.CompletionItem, error)
	HandleSignatureHelp(params lsp.TextDocumentPositionParams, conn Connection) (*lsp.SignatureHelp, error)
	HandleFormatting(params lsp.DocumentFormattingParams, conn Connection) ([]*lsp.TextEdit, error)
	// HandleRangeFormatting
	HandleHighlights(params lsp.DocumentHighlightParams, conn Connection) ([]*lsp.DocumentHighlight, error)
	HandleCodeAction(params lsp.CodeActionParams, conn Connection) (*lsp.CodeAction, error)
	HandleCodeActionResolve(params lsp.CodeAction, conn Connection) (*lsp.CodeAction, error)
	HandleRename(params lsp.RenameParams) (*lsp.WorkspaceEdit, error)
	HandleSemanticTokensFull(params lsp.SemanticTokensParams, conn Connection) (*lsp.SemanticTokens, error)
	//HandlePrepareRename(params lsp.PrepareRenameParams) (*lsp.PrepareRenameResult, error)
	//HandleFoldingRange(params lsp.FoldingRangeParams) ([]*lsp.FoldingRange, error)
	//HandleSelectionRange(params lsp.SelectionRangeParams) ([]*lsp.SelectionRange, error)

	//HandlePrepareCallHierarchy(params lsp.CallHierarchyPrepareParams) ([]*lsp.CallHierarchyItem, error)
	//HandleCallHierarchyIncomingCalls(params lsp.CallHierarchyIncomingCallsParams, conn Connection) ([]*CallHierarchyIncomingCall, error)
	//HandleCallHierarchyOutgoingCalls(params lsp.CallHierarchyOutgoingCallsParams, conn Connection) ([]*CallHierarchyOutgoingCall, error)
	//HandleSemanticTokens(params lsp.SemanticTokensParams) (*lsp.SemanticTokens, error)

	//HandleMonikers()

	// HandleLink
	// HandleLinkResolve
	// HandleColor
	// HandleColorPresentation

	/**
	 * A code lens represents a command that should be shown along with
	 * source text, like the number of references, a way to run tests, etc.
	 *
	 * A code lens is _unresolved_ when no command is associated to it. For
	 * performance reasons the creation of a code lens and resolving should be done
	 * in two stages.
	 */
	HandleCodeLens(params lsp.CodeLensParams, conn Connection) ([]*lsp.CodeLens, error)
	HandleCodeLensResolve(params lsp.CodeLens, conn Connection) (*lsp.CodeLens, error)

	// File System
	HandleDidChangeWatchedFiles(params lsp.DidChangeWatchedFilesParams, conn Connection) error
	HandleDidOpen(params lsp.DidOpenTextDocumentParams, conn Connection) error
	HandleDidChange(params lsp.DidChangeTextDocumentParams, conn Connection) error
	HandleDidClose(params lsp.DidCloseTextDocumentParams, conn Connection) error
	HandleWillSave(params lsp.WillSaveTextDocumentParams, conn Connection) error
	HandleDidSave(params lsp.DidSaveTextDocumentParams, conn Connection) error
}

type SendOut struct {
	conn jsonrpc2.JSONRPC2
	ctx  context.Context
}

func NewSendOut(conn jsonrpc2.JSONRPC2, ctx context.Context) *SendOut {
	return &SendOut{conn: conn, ctx: ctx}
}

func (s *SendOut) PublishDiagnostics(params lsp.PublishDiagnosticsParams) error {
	return s.conn.Notify(s.ctx, "textDocument/publishDiagnostics", params)
}

type HandleLspRequests struct {
	handler       Handler
	isInitialized bool
}

func NewLspRequests(handler Handler) *HandleLspRequests {
	return &HandleLspRequests{handler: handler}
}

func isFileSystemRequest(method string) bool {
	return method == "textDocument/didOpen" ||
		method == "textDocument/didChange" ||
		method == "textDocument/willSave" ||
		method == "textDocument/didSave" ||
		method == "textDocument/didClose"
}

func (h *HandleLspRequests) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	result, err := h.HandleInternal(ctx, conn, req)
	if err != nil {
		log.Printf("error: %v\n", err)
	}
	if req.Notif {
		if err != nil {
			log.Printf("HandleLspRequests: notification %q error: %v\n", req.Method, err)
		}
		return
	}

	resp := &jsonrpc2.Response{ID: req.ID}

	if err == nil {
		err = resp.SetResult(result)
	}

	if err != nil {
		log.Printf("HandleLspRequests: json response error %v\n", err)

		if e, ok := err.(*jsonrpc2.Error); ok {
			resp.Error = e
		} else {
			resp.Error = &jsonrpc2.Error{Message: err.Error()}
		}
	}

	if !req.Notif {
		if err := conn.SendResponse(ctx, resp); err != nil {
			if err != jsonrpc2.ErrClosed {
				log.Printf("HandleLspRequests: sending response %s: %v\n", resp.ID, err)
			}
		}
	}
}

func (h *HandleLspRequests) handleFileSystemRequest(ctx context.Context, req *jsonrpc2.Request, conn Connection) error {
	switch req.Method {
	case "textDocument/didOpen":
		var params lsp.DidOpenTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return err
		}
		return h.handler.HandleDidOpen(params, conn)

	case "textDocument/didChange":
		var params lsp.DidChangeTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return err
		}

		return h.handler.HandleDidChange(params, conn)

	case "textDocument/didClose":
		var params lsp.DidCloseTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return err
		}
		return h.handler.HandleDidClose(params, conn)

	case "textDocument/willSave":
		var params lsp.WillSaveTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return err
		}
		return h.handler.HandleWillSave(params, conn)

	case "textDocument/didSave":
		var params lsp.DidSaveTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return err
		}
		return h.handler.HandleDidSave(params, conn)

	default:
		return fmt.Errorf("HandleLspRequests: unexpected file system request %v ", req.Method)
	}
}

func (h *HandleLspRequests) HandleInternal(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Method != "initialize" && !h.isInitialized {
		return nil, errors.New("HandleLspRequests: language server must be initialized, before issuing any other commands")
	}

	out := NewSendOut(conn, ctx)

	switch req.Method {
	case "initialize":
		if h.isInitialized {
			return nil, errors.New("HandleLspRequests: language server has already been initialized")
		}

		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Data: nil, Message: ""}
		}

		if err := h.handler.Reset(); err != nil {
			return nil, fmt.Errorf("reset failed %w", err)
		}

		h.isInitialized = true

		kind := lsp.TDSKIncremental

		tokenTypes := []string{
			"namespace",
			"type",
			"class",
			"enum",
			"interface",
			"struct",
			"typeParameter",
			"parameter",
			"variable",
			"property",
			"enumMember",
			"event",
			"function",
			"method",
			"macro",
			"keyword",
			"modifier",
			"comment",
			"string",
			"number",
			"regexp",
			"operator",
		}
		tokenModifiers := []string{
			"declaration",
			"definition",
			"readonly",
			"static",
			"deprecated",
			"abstract",
			"async",
			"modification",
			"documentation",
			"defaultLibrary",
		}
		syncOptions := lsp.TextDocumentSyncOptions{
			OpenClose:         true,
			Change:            kind,
			WillSave:          true,
			WillSaveWaitUntil: false,
			Save:              &lsp.SaveOptions{IncludeText: true},
		}

		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				TextDocumentSync:       &lsp.TextDocumentSyncOptionsOrKind{Options: &syncOptions, Kind: nil},
				CompletionProvider:     &lsp.CompletionOptions{ResolveProvider: false, TriggerCharacters: []string{"."}},
				HoverProvider:          true,
				SignatureHelpProvider:  &lsp.SignatureHelpOptions{TriggerCharacters: []string{"(", ","}},
				DeclarationProvider:    nil,
				DefinitionProvider:     true,
				TypeDefinitionProvider: true,
				ImplementationProvider: &lsp.ImplementationOptions{},
				ReferencesProvider: &lsp.ReferenceOptions{
					WorkDoneProgressOptions: lsp.WorkDoneProgressOptions{
						WorkDoneProgress: false,
					},
				},
				DocumentHighlightProvider:       &lsp.DocumentHighlightOptions{},
				DocumentSymbolProvider:          true,
				DocumentLinkProvider:            nil, // TODO: Not sure what this is yet.
				ColorProvider:                   nil,
				DocumentFormattingProvider:      true,
				CodeActionProvider:              false,
				CodeLensProvider:                &lsp.CodeLensOptions{ResolveProvider: false},
				DocumentRangeFormattingProvider: false,
				DocumentOnTypeFormattingProvider: &lsp.DocumentOnTypeFormattingOptions{
					FirstTriggerCharacter: "",
					MoreTriggerCharacter:  []string{},
				},
				RenameProvider:       true,
				FoldingRangeProvider: nil,
				ExecuteCommandProvider: &lsp.ExecuteCommandOptions{
					Commands: []string{},
				},
				SelectionRangeProvider: &lsp.SelectionRangeOptions{
					WorkDoneProgressOptions: lsp.WorkDoneProgressOptions{
						WorkDoneProgress: false,
					},
				},
				LinkedEditingRangeProvider: &lsp.LinkedEditingRangeOptions{
					WorkDoneProgressOptions: lsp.WorkDoneProgressOptions{
						WorkDoneProgress: false,
					},
				},
				CallHierarchyProvider: &lsp.CallHierarchyOptions{
					WorkDoneProgressOptions: lsp.WorkDoneProgressOptions{
						WorkDoneProgress: false,
					},
				},
				SemanticTokensProvider: &lsp.SemanticTokensOptions{
					WorkDoneProgressOptions: lsp.WorkDoneProgressOptions{
						WorkDoneProgress: false,
					},
					Legend: lsp.SemanticTokensLegend{
						TokenTypes:     tokenTypes,
						TokenModifiers: tokenModifiers,
					},
					Range: false,
					Full: &lsp.SemanticTokenOptionsFull{
						Delta: false,
					},
				},
				MonikerProvider: &lsp.MonikerOptions{
					WorkDoneProgressOptions: lsp.WorkDoneProgressOptions{
						WorkDoneProgress: false,
					},
				},
				WorkspaceSymbolProvider: true,
				Workspace: &lsp.WorkspaceOptions{
					WorkspaceFolders: &lsp.WorkspaceFoldersServerCapabilities{
						Supported:           false,
						ChangeNotifications: "",
					},
					FileOperations: &lsp.WorkspaceOptionsFileOperations{
						DidCreate: &lsp.FileOperationRegistrationOptions{
							Filters: []lsp.FileOperationFilter{},
						},
						WillCreate: &lsp.FileOperationRegistrationOptions{
							Filters: []lsp.FileOperationFilter{},
						},
						DidRename: &lsp.FileOperationRegistrationOptions{
							Filters: []lsp.FileOperationFilter{},
						},
						WillRename: &lsp.FileOperationRegistrationOptions{
							Filters: []lsp.FileOperationFilter{},
						},
						DidDelete: &lsp.FileOperationRegistrationOptions{
							Filters: []lsp.FileOperationFilter{},
						},
						WillDelete: &lsp.FileOperationRegistrationOptions{
							Filters: []lsp.FileOperationFilter{},
						},
					},
				},
				Experimental:                 nil,
				XWorkspaceReferencesProvider: false,
				XDefinitionProvider:          false,
				XWorkspaceSymbolByProperties: false,
			},
		}, nil

	case "initialized":
		// A notification that the client is ready to receive requests. TODO: should check client capabilities
		return nil, nil

	case "shutdown":
		h.handler.ShutDown()

		return nil, nil

	case "exit":
		if c, ok := conn.(*jsonrpc2.Conn); ok {
			c.Close()
		}

		return nil, nil

	case "$/cancelRequest":
		if req.Params == nil {
			return nil, nil
		}

		var params lsp.CancelParams

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, nil
		}

		return nil, nil

	case "textDocument/hover":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.TextDocumentPositionParams

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		return h.handler.HandleHover(params, out)

	case "textDocument/definition":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.TextDocumentPositionParams

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handler.HandleGotoDefinition(params, out)

	case "textDocument/declaration":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.DeclarationOptions

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		return h.handler.HandleGotoDeclaration(params, out)

	case "textDocument/typeDefinition":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		return h.handler.HandleGotoTypeDefinition(params, out)

	case "textDocument/completion":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.CompletionParams

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		return h.handler.HandleCompletion(params, out)
	case "completionItem/resolve":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.CompletionItem

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		return h.handler.HandleCompletionItemResolve(params, out)
	case "textDocument/references":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.ReferenceParams

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		return h.handler.HandleFindReferences(params, out)
	case "textDocument/implementation":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.TextDocumentPositionParams

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handler.HandleGotoImplementation(params, out)
	case "textDocument/documentSymbol":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.DocumentSymbolParams

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		return h.handler.HandleSymbol(params, out)
	case "textDocument/linkedEditingRange":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.LinkedEditingRangeParams

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		return h.handler.HandleLinkedEditingRange(params, out)

	case "textDocument/semanticTokens/full":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.SemanticTokensParams

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		return h.handler.HandleSemanticTokensFull(params, out)

	case "textDocument/signatureHelp":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.TextDocumentPositionParams

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		return h.handler.HandleSignatureHelp(params, out)

	case "textDocument/formatting":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.DocumentFormattingParams

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		return h.handler.HandleFormatting(params, out)

	case "textDocument/codeAction":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.CodeActionParams

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		return h.handler.HandleCodeAction(params, out)

	case "textDocument/documentHighlight":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.DocumentHighlightParams

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		return h.handler.HandleHighlights(params, out)

	case "textDocument/codeLens":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.CodeLensParams

		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handler.HandleCodeLens(params, out)

	case "workspace/didChangeWatchedFiles":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}

		var params lsp.DidChangeWatchedFilesParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		return nil, h.handler.HandleDidChangeWatchedFiles(params, out)

	default:
		if isFileSystemRequest(req.Method) {
			err := h.handleFileSystemRequest(ctx, req, out)

			return nil, err
		}

		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("HandleLspRequests: request is not supported: %s", req.Method)}
	}
}
