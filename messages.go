package deadweight

import (
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
)

type message struct {
	ID     int32           `json:"id"`
	Result json.RawMessage `json:"result"`
}

type messageHandler func(message)

func initializeResponse(ready chan struct{}) messageHandler {
	return func(m message) {
		ready <- struct{}{}
	}
}

type range_ struct {
	Start Position `json:"start"`
}

type documentSymbol struct {
	Name           string           `json:"name"`
	Kind           int              `json:"kind"`
	SelectionRange range_           `json:"selectionRange"`
	Children       []documentSymbol `json:"children"`
}

func documentSymbolResponse(wg *sync.WaitGroup, symbols *SymbolMap, filePath string) messageHandler {
	return func(m message) {
		defer wg.Done()
		var results []documentSymbol
		if err := json.Unmarshal(m.Result, &results); err != nil {
			slog.Error("document symbol response unmarshal error", slog.Any("error", err))
			return
		}
		if len(results) == 0 {
			return
		}
		var fileSymbols []Symbol
		for _, result := range results {
			fileSymbols = append(fileSymbols, Symbol{
				Position: result.SelectionRange.Start,
				Name:     result.Name,
				Kind:     result.Kind,
			})
			for _, child := range result.Children {
				fileSymbols = append(fileSymbols, Symbol{
					Position: child.SelectionRange.Start,
					Name:     child.Name,
					Kind:     child.Kind,
				})
			}
		}
		symbols.Store(filePath, fileSymbols)
	}
}

type reference struct {
	URI string `json:"uri"`
}

func isUsed(references []reference) bool {
	for _, reference := range references {
		if strings.HasSuffix(reference.URI, "_test.go") || strings.Contains(reference.URI, "mock") {
			continue
		}
		return true
	}
	return false
}

func referencesResponse(wg *sync.WaitGroup, filePath string, symbol Symbol, unusedSymbols *SymbolMap) messageHandler {
	return func(m message) {
		defer wg.Done()

		var references []reference
		if err := json.Unmarshal(m.Result, &references); err != nil {
			slog.Error("references response unmarshal error", slog.Any("error", err))
			return
		}

		if !isUsed(references) && symbol.Name != "main" {
			unusedSymbols.Add(filePath, symbol)
		}
	}
}
