package lspserv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/sourcegraph/go-langserver/langserver"
	"github.com/sourcegraph/go-langserver/langserver/util"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type Handler interface {
	Reset(init *langserver.InitializeParams) error
	ShutDown()
	HandleHover(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (*lsp.Hover, error)
	ResetCaches(lock bool)
}

type HandleLspRequests struct {
	handler Handler
	mu      sync.Mutex
	init    *langserver.InitializeParams // set by "initialize" request

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
	log.Printf("got an  json request! %v\n", req)
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

	log.Printf("got an  json response to send! %v\n", result)

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
		log.Printf("sending response! %v\n", resp)
		if err := conn.SendResponse(ctx, resp); err != nil {
			if err != jsonrpc2.ErrClosed {
				log.Printf("jsonrpc2 handler: sending response %s: %v\n", resp.ID, err)
			}
		}
	} else {
		log.Printf("didn't care to send response")
	}
}

func (h *HandleLspRequests) HandleInternal(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request) (result interface{}, err error) {
	log.Printf("got an request! %v\n", req)
	// Prevent any uncaught panics from taking the entire server down.
	defer func() {
		if perr := util.Panicf(recover(), "%v", req.Method); perr != nil {
			err = perr
		}
	}()

	h.mu.Lock()
	if req.Method != "initialize" && h.init == nil {
		h.mu.Unlock()
		return nil, errors.New("server must be initialized")
	}
	h.mu.Unlock()

	switch req.Method {
	case "initialize":
		if h.init != nil {
			return nil, errors.New("language server is already initialized")
		}
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params langserver.InitializeParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}

		// HACK: RootPath is not a URI, but historically we treated it
		// as such. Convert it to a file URI
		//if params.RootPath != "" && !util.IsURI(lsp.DocumentURI(params.RootPath)) {
		params.RootPath = string(util.PathToURI(params.RootPath))
		//}

		if err := h.handler.Reset(&params); err != nil {
			return nil, err
		}

		h.init = &params

		// PERF: Kick off a workspace/symbol in the background to warm up the server
		kind := lsp.TDSKIncremental
		var completionOp *lsp.CompletionOptions

		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
					Kind: &kind,
				},
				CompletionProvider:           completionOp,
				DefinitionProvider:           true,
				TypeDefinitionProvider:       true,
				DocumentFormattingProvider:   true,
				DocumentSymbolProvider:       true,
				HoverProvider:                true,
				ReferencesProvider:           true,
				WorkspaceSymbolProvider:      true,
				ImplementationProvider:       true,
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
		return h.handler.HandleHover(ctx, conn, req, params)

		/*
			case "textDocument/definition":
				if req.Params == nil {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
				}
				var params lsp.TextDocumentPositionParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
				return h.handleDefinition(ctx, conn, req, params)

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

			case "textDocument/completion":
				if !h.config.GocodeCompletionEnabled {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound,
						Message: fmt.Sprintf("completion is disabled. Enable with flag `-gocodecompletion`")}
				}
				if req.Params == nil {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
				}
				var params lsp.CompletionParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
				return h.handleTextDocumentCompletion(ctx, conn, req, params)

			case "textDocument/references":
				if req.Params == nil {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
				}
				var params lsp.ReferenceParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
				return h.handleTextDocumentReferences(ctx, conn, req, params)

			case "textDocument/implementation":
				if req.Params == nil {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
				}
				var params lsp.TextDocumentPositionParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
				return h.handleTextDocumentImplementation(ctx, conn, req, params)

			case "textDocument/documentSymbol":
				if req.Params == nil {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
				}
				var params lsp.DocumentSymbolParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
				return h.handleTextDocumentSymbol(ctx, conn, req, params)

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
			/*
				uri, fileChanged, err := h.handler.handleFileSystemRequest(ctx, req)
				if fileChanged {
					// a file changed, so we must re-typecheck and re-enumerate symbols
					h.handler.resetCaches(true)
				}
				if uri != "" {
					// a user is viewing this path, hint to add it to the cache
					// (unless we're primarily using binary package cache .a
					// files).
				}
			*/
			return nil, err
		}

		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
	}
}
