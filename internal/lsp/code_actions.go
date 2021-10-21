package lsp

import (
	"sort"

	lsp "github.com/hashicorp/terraform-ls/internal/protocol"
)

const (
	SourceFormatAll            = "source.formatAll"
	SourceFormatAllTerraformLs = "source.formatAll.terraform-ls"
)

type CodeActions map[lsp.CodeActionKind]bool

var (
	SupportedCodeActions = CodeActions{
		// `source.*`: Source code actions apply to the entire file. They must be explicitly
		// requested and will not show in the normal lightbulb menu. Source actions
		// can be run on save using editor.codeActionsOnSave and are also shown in
		// the source context menu.
		// For action definitions, refer to: https://code.visualstudio.com/api/references/vscode-api#CodeActionKind

		// `source.fixAll`: Fix all actions automatically fix errors that have a clear fix that do
		// not require user input. They should not suppress errors or perform unsafe
		// fixes such as generating new types or classes.
		// ** We don't support this as terraform fmt only adjusts style**
		// lsp.SourceFixAll: true,

		// `source.formatAll`: Generic format code action. We register this as
		// terraform fmt can apply to entire documents or ranges
		SourceFormatAll: true,

		// `source.formatAll.terraform-ls`: Terraform specific format code action.
		// We register so that users can choose to allow only terraform type formatting
		SourceFormatAllTerraformLs: true,
	}
)

func (c CodeActions) AsSlice() []lsp.CodeActionKind {
	s := make([]lsp.CodeActionKind, 0)
	for v := range c {
		s = append(s, v)
	}

	sort.SliceStable(s, func(i, j int) bool {
		return string(s[i]) < string(s[j])
	})
	return s
}

func (ca CodeActions) Only(only []lsp.CodeActionKind) CodeActions {
	wanted := make(CodeActions, 0)

	for _, kind := range only {
		if v, ok := ca[kind]; ok {
			wanted[kind] = v
		}
	}

	return wanted
}
