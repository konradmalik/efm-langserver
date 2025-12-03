package lsp

import (
	"context"
	"encoding/json"

	"github.com/konradmalik/flint-ls/types"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LspHandler) HandleTextDocumentDidChange(_ context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params types.DidChangeTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	for _, change := range params.ContentChanges {
		if err := h.langHandler.UpdateFile(params.TextDocument.URI, change.Text, &params.TextDocument.Version); err != nil {
			return nil, err
		}
	}

	notifier := NewNotifier(conn)
	h.ScheduleLinting(*notifier, params.TextDocument.URI, types.EventTypeChange)

	return nil, nil
}
