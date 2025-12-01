package deadweight

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Symbol struct {
	Position Position `json:"position"`
	Name     string   `json:"name"`
	Kind     int      `json:"kind"`
}

type SymbolMap struct {
	m map[string][]Symbol

	sync.Mutex
}

type Test struct {
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

func (sm *SymbolMap) ReferencesSymbols(ctx context.Context, lc *lspClient) (*SymbolMap, error) {
	defer sm.Unlock()

	wg := &sync.WaitGroup{}
	unusedSymbols := NewSymbolMap()

	sm.Lock()
	for filePath, symbols := range sm.m {
		for _, symbol := range symbols {
			wg.Add(1)
			if err := lc.References(ctx, wg,
				filePath,
				symbol,
				unusedSymbols,
			); err != nil {
				return nil, err
			}
		}
	}
	wg.Wait()

	return unusedSymbols, nil
}
