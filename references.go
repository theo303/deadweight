package deadweight

import (
	"strings"
	"sync"

	"github.com/theo303/deadweight/lsp"
)

type ReferenceMap struct {
	m map[string]map[Symbol][]lsp.Location

	sync.Mutex
}

func NewReferenceMap() *ReferenceMap {
	return &ReferenceMap{
		m: make(map[string]map[Symbol][]lsp.Location),
	}
}

func (rm *ReferenceMap) Store(filePath string, symbol Symbol, referencesURIs []lsp.Location) {
	defer rm.Unlock()
	rm.Lock()
	if rm.m[filePath] == nil {
		rm.m[filePath] = make(map[Symbol][]lsp.Location)
	}
	rm.m[filePath][symbol] = referencesURIs
}

func (rm *ReferenceMap) GetUnusedSymbols() *SymbolMap {
	unusedSymbols := NewSymbolMap()

	for filePath, symbols := range rm.m {
		for symbol, references := range symbols {
			if !isUsed(references) && symbol.Name != "main" {
				unusedSymbols.Add(filePath, symbol)
			}
		}
	}

	return unusedSymbols
}

func isUsed(references []lsp.Location) bool {
	for _, reference := range references {
		if strings.HasSuffix(reference.URI, "_test.go") || strings.Contains(reference.URI, "mock") {
			continue
		}
		return true
	}
	return false
}
