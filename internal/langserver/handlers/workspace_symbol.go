package handlers

import (
	"context"

	lsctx "github.com/hashicorp/terraform-ls/internal/context"
	"github.com/hashicorp/terraform-ls/internal/decoder"
	ilsp "github.com/hashicorp/terraform-ls/internal/lsp"
	lsp "github.com/hashicorp/terraform-ls/internal/protocol"
)

func (svc *service) WorkspaceSymbol(ctx context.Context, params lsp.WorkspaceSymbolParams) ([]lsp.SymbolInformation, error) {
	cc, err := lsctx.ClientCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	d := decoder.NewDecoder(ctx, &decoder.PathReader{
		ModuleReader: svc.modStore,
		SchemaReader: svc.schemaStore,
	})

	symbols, err := d.Symbols(ctx, params.Query)
	if err != nil {
		return nil, err
	}

	return ilsp.WorkspaceSymbols(symbols, cc.Workspace.Symbol), nil
}
