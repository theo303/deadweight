package deadweight

import (
	"path/filepath"
	"slices"

	"github.com/theo303/deadweight/lsp"
)

type Rules struct {
	IgnoreRules          []IgnoreRule
	IgnoreEmbeddedFields bool
}

func (r Rules) KeepSymbol(filePath string, s Symbol) bool {
	if s.IsEmbeddedField && r.IgnoreEmbeddedFields {
		return false
	}
	for _, ir := range r.IgnoreRules {
		if ir.ignore(filePath, s) {
			return false
		}
	}
	return true
}

type IgnoreRule struct {
	Kinds []lsp.SymbolKind
	Names []string
}

func (ir IgnoreRule) ignore(_ string, s Symbol) bool {
	if !slices.Contains(ir.Kinds, s.Kind) {
		return false
	}
	for _, filter := range ir.Names {
		if match, _ := filepath.Match(filter, s.Name); match {
			return true
		}
	}
	return false
}
