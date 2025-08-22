package lsp

import (
	"context"
	"encoding/json"

	"github.com/konradmalik/efm-langserver/types"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LspHandler) HandleTextDocumentDidSave(_ context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params types.DidSaveTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	var event types.EventType
	if params.Text != nil {
		err = h.langHandler.UpdateFile(params.TextDocument.URI, *params.Text, nil)
		event = types.EventTypeSave
	} else {
		event = types.EventTypeChange
	}
	if err != nil {
		return nil, err
	}

	notifier := NewNotifier(conn)
	h.ScheduleLinting(*notifier, params.TextDocument.URI, event)

	return nil, nil
}
