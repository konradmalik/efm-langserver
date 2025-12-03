package lsp

import (
	"context"
	"encoding/json"

	"github.com/konradmalik/flint-ls/logs"
	"github.com/konradmalik/flint-ls/types"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LspHandler) HandleWorkspaceDidChangeConfiguration(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params types.DidChangeConfigurationParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	logs.Log.SetLevel(logs.LogLevel(params.Settings.LogLevel))
	h.UpdateConfiguration(&params.Settings)
	return nil, nil
}
