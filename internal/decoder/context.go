package decoder

import (
	"context"
	"fmt"

	ilsp "github.com/hashicorp/terraform-ls/internal/lsp"
)

type clientNameCtxKey struct{}

func ContextWithClientName(ctx context.Context, namePtr *string) context.Context {
	return context.WithValue(ctx, clientNameCtxKey{}, namePtr)
}

func ClientName(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(clientNameCtxKey{}).(*string)
	if !ok {
		return "", false
	}
	return *name, true
}

func SetClientName(ctx context.Context, name string) error {
	namePtr, ok := ctx.Value(clientNameCtxKey{}).(*string)
	if !ok {
		return fmt.Errorf("missing context: client name")
	}

	*namePtr = name
	return nil
}

type languageIdCtxKey struct{}

func WithLanguageId(ctx context.Context, langId ilsp.LanguageID) context.Context {
	return context.WithValue(ctx, languageIdCtxKey{}, langId)
}

func LanguageId(ctx context.Context) (ilsp.LanguageID, bool) {
	id, ok := ctx.Value(languageIdCtxKey{}).(ilsp.LanguageID)
	if !ok {
		return "", false
	}
	return id, true
}
