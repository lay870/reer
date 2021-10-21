package handlers

import (
	"context"

	lsctx "github.com/hashicorp/terraform-ls/internal/context"
	ilsp "github.com/hashicorp/terraform-ls/internal/lsp"
	lsp "github.com/hashicorp/terraform-ls/internal/protocol"
)

func (svc *service) GoToReferenceTarget(ctx context.Context, params lsp.TextDocumentPositionParams) (interface{}, error) {
	cc, err := lsctx.ClientCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	fs, err := lsctx.DocumentStorage(ctx)
	if err != nil {
		return nil, err
	}

	doc, err := fs.GetDocument(ilsp.FileHandlerFromDocumentURI(params.TextDocument.URI))
	if err != nil {
		return nil, err
	}

	d, err := svc.decoderForDocument(ctx, doc)
	if err != nil {
		return nil, err
	}

	fPos, err := ilsp.FilePositionFromDocumentPosition(params, doc)
	if err != nil {
		return nil, err
	}

	svc.logger.Printf("Looking for ref origin at %q -> %#v", doc.Filename(), fPos.Position())
	origin, err := d.ReferenceOriginAtPos(doc.Filename(), fPos.Position())
	if err != nil {
		return nil, err
	}
	if origin == nil {
		return nil, nil
	}
	svc.logger.Printf("found origin: %#v", origin)

	target, err := d.ReferenceTargetForOrigin(*origin)
	if err != nil {
		return nil, err
	}

	return ilsp.ReferenceToLocationLink(doc.Dir(), *origin, target, cc.TextDocument.Declaration.LinkSupport), nil
}
