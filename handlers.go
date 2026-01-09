package deadweight

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/theo303/deadweight/lsp"
)

type messageHandler func(lsp.Message)

func initializeResponse(ready chan struct{}) messageHandler {
	return func(m lsp.Message) {
		ready <- struct{}{}
	}
}

func getAllSymbols(documentSymbol lsp.DocumentSymbol) []lsp.DocumentSymbol {
	children := make([]lsp.DocumentSymbol, 0, len(documentSymbol.Children))
	children = append(children, documentSymbol)
	for _, child := range documentSymbol.Children {
		children = append(children, getAllSymbols(child)...)
	}
	return children
}

func (lc *lspClient) documentSymbolResponse(wg *sync.WaitGroup, symbols *SymbolMap, filePath string) messageHandler {
	return func(m lsp.Message) {
		defer wg.Done()
		var results []lsp.DocumentSymbol
		if err := json.Unmarshal(m.Result, &results); err != nil {
			slog.Error("document symbol response unmarshal error", slog.Any("error", err))
			return
		}
		if len(results) == 0 {
			return
		}
		var fileSymbols []Symbol
		var err error
		for _, result := range results {
			for _, symbol := range getAllSymbols(result) {
				s := NewSymbol(symbol)
				s.IsEmbeddedField, err = lc.isEmbedded(filePath, symbol)
				if err != nil {
					slog.Error("isEmbedded error, skipping symbol", slog.Any("error", err),
						slog.String("filePath", filePath),
						slog.String("symbolName", symbol.Name),
						slog.Int("symbolLine", symbol.SelectionRange.Start.Line),
						slog.Int("symbolCharacter", symbol.SelectionRange.Start.Character),
					)
					continue
				}
				if lc.rules.KeepSymbol(filePath, s) {
					fileSymbols = append(fileSymbols, s)
				}
			}
		}
		symbols.Store(filePath, fileSymbols)
	}
}

func referencesResponse(wg *sync.WaitGroup, references *ReferenceMap, filePath string, symbol Symbol) messageHandler {
	return func(m lsp.Message) {
		defer wg.Done()

		var symbolReferences []lsp.Location
		if err := json.Unmarshal(m.Result, &symbolReferences); err != nil {
			slog.Error("references response unmarshal error", slog.Any("error", err))
			return
		}

		references.Store(filePath, symbol, symbolReferences)
	}
}

func prepareTypeHierarchyResponse(hasParentType chan bool) messageHandler {
	return func(m lsp.Message) {
		hasParentType <- len(m.Result) != 0
	}
}
