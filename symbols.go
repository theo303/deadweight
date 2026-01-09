package deadweight

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/theo303/deadweight/lsp"
)

type Symbol struct {
	Position lsp.Position
	Name     string
	Kind     lsp.SymbolKind

	IsEmbeddedField bool
}

func NewSymbol(documentSymbol lsp.DocumentSymbol) Symbol {
	return Symbol{
		Position: documentSymbol.SelectionRange.Start,
		Name:     documentSymbol.Name,
		Kind:     documentSymbol.Kind,
	}
}

type SymbolMap struct {
	m map[string][]Symbol

	sync.Mutex
}

func NewSymbolMap() *SymbolMap {
	return &SymbolMap{
		m: make(map[string][]Symbol),
	}
}

func (sm *SymbolMap) Store(filePath string, symbols []Symbol) {
	defer sm.Unlock()
	sm.Lock()
	sm.m[filePath] = symbols
}

func (sm *SymbolMap) Add(filepath string, symbol Symbol) {
	defer sm.Unlock()
	sm.Lock()
	if sm.m[filepath] == nil {
		sm.m[filepath] = make([]Symbol, 0)
	}
	sm.m[filepath] = append(sm.m[filepath], symbol)
}

func (sm *SymbolMap) Print() {
	defer sm.Unlock()

	sm.Lock()
	for filePath, symbols := range sm.m {
		for _, symbol := range symbols {
			slog.Info(fmt.Sprintf("%s (%s) %s:%d:%d",
				symbol.Name, symbol.Kind.String(), filePath, symbol.Position.Line+1, symbol.Position.Character+1,
			))
		}
	}
}
