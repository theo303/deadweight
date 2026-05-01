package deadweight

import (
	"fmt"

	"github.com/theo303/deadweight/lsp"
)

type Config struct {
	IgnoreSymbols        []ignoreSymbolsConfig `yaml:"ignore-symbols"`
	IgnoreEmbeddedFields bool                  `yaml:"ignore-embedded-fields"`
}

type ignoreSymbolsConfig struct {
	Kinds []string `yaml:"kinds"`
	Names []string `yaml:"names"`
}

func (c Config) ToRules() (Rules, error) {
	ignoreSymbols := make([]IgnoreSymbols, 0, len(c.IgnoreSymbols))
	for _, isc := range c.IgnoreSymbols {
		var kinds []lsp.SymbolKind
		for _, symbolName := range isc.Kinds {
			sk, err := lsp.ParseSymbolKind(symbolName)
			if err != nil {
				return Rules{}, fmt.Errorf("invalid symbol kind '%s': %w", symbolName, err)
			}
			kinds = append(kinds, sk)
		}

		ignoreSymbols = append(ignoreSymbols, IgnoreSymbols{
			Kinds: kinds,
			Names: isc.Names,
		})
	}
	return Rules{
		ignoreSymbols:        ignoreSymbols,
		ignoreEmbeddedFields: c.IgnoreEmbeddedFields,
	}, nil
}
