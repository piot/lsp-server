package lspserv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"sync"

	"github.com/piot/go-lsp"
	"github.com/piot/jsonrpc2"
)

type Connection interface {
	PublishDiagnostics(params lsp.PublishDiagnosticsParams) error
}


type Handler interface {
	Reset() error
	ResetCaches(lock bool)
	ShutDown()
	HandleHover(params lsp.TextDocumentPositionParams, conn Connection) (*lsp.Hover, error)
	HandleGotoDefinition(params lsp.TextDocumentPositionParams, conn Connection) (*lsp.Location, error)
	HandleTextDocumentReferences(params lsp.ReferenceParams, conn Connection) ([]*lsp.Location, error)
	HandleTextDocumentSymbol(params lsp.DocumentSymbolParams, conn Connection) ([]*lsp.DocumentSymbol, error) // Used for outline
	HandleTextDocumentCompletion(params lsp.CompletionParams, conn Connection) (*lsp.CompletionList, error) // Intellisense when pressing '.'.
}

type SendOut struct {
	conn jsonrpc2.JSONRPC2
	ctx context.Context
}

func NewSendOut(conn jsonrpc2.JSONRPC2, ctx context.Context) *SendOut {
	return &SendOut{conn: conn, ctx:ctx}
}

func (s *SendOut) PublishDiagnostics(params lsp.PublishDiagnosticsParams) error {
	return s.conn.Notify(s.ctx, "textDocument/publishDiagnostics", params)
}


type HandleLspRequests struct {
	handler Handler
	mu      sync.Mutex
	init    bool
}

func NewLspRequests(handler Handler) *HandleLspRequests {
	return &HandleLspRequests{handler: handler}
}

// isFileSystemRequest returns if this is an LSP method whose sole
// purpose is modifying the contents of the overlay file system.
func isFileSystemRequest(method string) bool {
	return method == "textDocument/didOpen" ||
		method == "textDocument/didChange" ||
		method == "textDocument/didClose" ||
		method == "textDocument/didSave"
}

// handle implements jsonrpc2.Handler.
func (h *HandleLspRequests) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) /*(result interface{}, err error)*/ {
	result, err := h.HandleInternal(ctx, conn, req)
	if err != nil {
		log.Printf("error: %v\n", err)
	}
	if req.Notif {
		if err != nil {
			log.Printf("jsonrpc2 handler: notification %q handling error: %v\n", req.Method, err)
		}
		return
	}

	resp := &jsonrpc2.Response{ID: req.ID}
	if err == nil {
		err = resp.SetResult(result)
	}
	if err != nil {
		log.Printf("got an  json response was an error %v\n", err)
		if e, ok := err.(*jsonrpc2.Error); ok {
			resp.Error = e
		} else {
			resp.Error = &jsonrpc2.Error{Message: err.Error()}
		}
	}

	if !req.Notif {
		if err := conn.SendResponse(ctx, resp); err != nil {
			if err != jsonrpc2.ErrClosed {
				log.Printf("jsonrpc2 handler: sending response %s: %v\n", resp.ID, err)
			}
		}
	} else {
		log.Printf("didn't care to send response")
	}
}

func (h *HandleLspRequests) readFile(ctx context.Context, uri lsp.DocumentURI) ([]byte, error) {
	url, err := url.Parse(string(uri))
	if err != nil {
		return nil, err
	}
	path := url.Path
	contents, err := ioutil.ReadFile(path)
	return contents, err
}

// handleFileSystemRequest handles textDocument/did* requests. The URI the
// request is for is returned. true is returned if a file was modified.
func (h *HandleLspRequests) handleFileSystemRequest(ctx context.Context, req *jsonrpc2.Request) (lsp.DocumentURI, bool, error) {
	do := func(uri lsp.DocumentURI, op func() error) (lsp.DocumentURI, bool, error) {
		before, beforeErr := h.readFile(ctx, uri)
		if beforeErr != nil && !os.IsNotExist(beforeErr) {
			// There is no op that could succeed in this case. (Most
			// commonly occurs when uri refers to a dir, not a file.)
			return uri, false, beforeErr
		}
		err := op()
		after, afterErr := h.readFile(ctx, uri)
		if os.IsNotExist(beforeErr) && os.IsNotExist(afterErr) {
			// File did not exist before or after so nothing has changed.
			return uri, false, err
		} else if afterErr != nil || beforeErr != nil {
			// If an error prevented us from reading the file
			// before or after then we assume the file changed to
			// be conservative.
			return uri, true, err
		}
		return uri, !bytes.Equal(before, after), err
	}

	switch req.Method {
	case "textDocument/didOpen":
		var params lsp.DidOpenTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return "", false, err
		}
		return do(params.TextDocument.URI, func() error {
			return nil
		})

	case "textDocument/didChange":
		var params lsp.DidChangeTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return "", false, err
		}
		return do(params.TextDocument.URI, func() error {
			return nil
		})

	case "textDocument/didClose":
		var params lsp.DidCloseTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return "", false, err
		}
		return do(params.TextDocument.URI, func() error {
			return nil
		})

	case "textDocument/didSave":
		var params lsp.DidSaveTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return "", false, err
		}
		// no-op
		return params.TextDocument.URI, false, nil

	default:
		panic("unexpected file system request method: " + req.Method)
	}
}

func (h *HandleLspRequests) HandleInternal(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request) (result interface{}, err error) {
	h.mu.Lock()
	if req.Method != "initialize" && !h.init {
		h.mu.Unlock()
		return nil, errors.New("server must be initialized")
	}
	h.mu.Unlock()

	out := NewSendOut(conn, ctx)

	switch req.Method {
	case "initialize":
		if h.init {
			return nil, errors.New("language server is already initialized")
		}
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		/*
			var params langserver.InitializeParams
			if err := json.Unmarshal(*req.Params, &params); err != nil {
				return nil, err
			}
		*/

		// HACK: RootPath is not a URI, but historically we treated it
		// as such. Convert it to a file URI
		//if params.RootPath != "" && !util.IsURI(lsp.DocumentURI(params.RootPath)) {
		//params.RootPath = string(util.PathToURI(params.RootPath))
		//}

		if err := h.handler.Reset(); err != nil {
			return nil, err
		}

		h.init = true

		// PERF: Kick off a workspace/symbol in the background to warm up the server
		kind := lsp.TDSKIncremental

		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
					Kind: &kind,
				},
				CompletionProvider:           &lsp.CompletionOptions{
					ResolveProvider:   false,
					TriggerCharacters: []string{"."},
				},
				DeclarationProvider: 		  &lsp.DeclarationOptions{},
				DefinitionProvider:           true,
				TypeDefinitionProvider:       true,
				DocumentFormattingProvider:   true,
				DocumentSymbolProvider:       true,
				HoverProvider:                true,
				ReferencesProvider:           true,
				WorkspaceSymbolProvider:      true,
				ImplementationProvider:       &lsp.ImplementationOptions{},
				XWorkspaceReferencesProvider: true,
				XDefinitionProvider:          true,
				XWorkspaceSymbolByProperties: true,
				SignatureHelpProvider:        &lsp.SignatureHelpOptions{TriggerCharacters: []string{"(", ","}},
			},
		}, nil

	case "initialized":
		// A notification that the client is ready to receive requests. Ignore
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
		// notification, don't send back results/errors
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
				/*

			case "textDocument/typeDefinition":
				if req.Params == nil {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
				}
				var params lsp.TextDocumentPositionParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
				return h.handleTypeDefinition(ctx, conn, req, params)

			case "textDocument/xdefinition":
				if req.Params == nil {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
				}
				var params lsp.TextDocumentPositionParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
				return h.handleXDefinition(ctx, conn, req, params)
*/
			case "textDocument/completion":
				if req.Params == nil {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
				}
				var params lsp.CompletionParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}

				return h.handler.HandleTextDocumentCompletion(params, out)
/*
			case "textDocument/references":
				if req.Params == nil {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
				}
				var params lsp.ReferenceParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
				return h.handler.HandleTextDocumentReferences(params, out)
/*
			case "textDocument/implementation":
				if req.Params == nil {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
				}
				var params lsp.TextDocumentPositionParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
				return h.handleTextDocumentImplementation(ctx, conn, req, params)
*/
			case "textDocument/documentSymbol":
				if req.Params == nil {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
				}
				var params lsp.DocumentSymbolParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
				return h.handler.HandleTextDocumentSymbol(params, out)
/*
			case "textDocument/signatureHelp":
				if req.Params == nil {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
				}
				var params lsp.TextDocumentPositionParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
				return h.handleTextDocumentSignatureHelp(ctx, conn, req, params)

			case "textDocument/formatting":
				if req.Params == nil {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
				}
				var params lsp.DocumentFormattingParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
				return h.handleTextDocumentFormatting(ctx, conn, req, params)

			case "workspace/symbol":
				if req.Params == nil {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
				}
				var params lspext.WorkspaceSymbolParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
				return h.handleWorkspaceSymbol(ctx, conn, req, params)

			case "workspace/xreferences":
				if req.Params == nil {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
				}
				var params lspext.WorkspaceReferencesParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
				return h.handleWorkspaceReferences(ctx, conn, req, params)
		*/

	default:
		if isFileSystemRequest(req.Method) {
			uri, fileChanged, err := h.handleFileSystemRequest(ctx, req)
			if fileChanged {
				// a file changed, so we must re-typecheck and re-enumerate symbols
				h.handler.ResetCaches(true)
			}
			log.Printf("fileystem change for '%v'\n", uri)
			return nil, err
		}

		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
	}
}
