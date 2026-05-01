package deadweight

import (
	"path/filepath"
	"slices"

	"github.com/theo303/deadweight/lsp"
)

type Rules struct {
	ignoreSymbols        []IgnoreSymbols
	ignoreEmbeddedFields bool
}

func (r Rules) KeepSymbol(filePath string, s Symbol) bool {
	if s.IsEmbeddedField && r.ignoreEmbeddedFields {
		return false
	}
	for _, ir := range r.ignoreSymbols {
		if ir.ignore(filePath, s) {
			return false
		}
	}
	return true
}

type IgnoreSymbols struct {
	Kinds []lsp.SymbolKind
	Names []string
}

func (ir IgnoreSymbols) ignore(_ string, s Symbol) bool {
	if !slices.Contains(ir.Kinds, s.Kind) {
		return false
	}
	if len(ir.Names) == 0 {
		return true
	}
	for _, filter := range ir.Names {
		if match, _ := filepath.Match(filter, s.Name); match {
			return true
		}
	}
	return false
}
