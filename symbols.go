package deadweight

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/theo303/deadweight/lsp"
)

type Symbol struct {
	Position lsp.Position
	Name     string
	Kind     int
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
		symbolsStr := make([]string, 0, len(symbols))
		for _, symbol := range symbols {
			symbolsStr = append(symbolsStr,
				fmt.Sprintf("%s %d:%d (%d)",
					symbol.Name, symbol.Position.Line, symbol.Position.Character, symbol.Kind,
				),
			)
		}
		slog.Info(filePath, slog.Any("symbols", strings.Join(symbolsStr, "; ")))
	}
}
