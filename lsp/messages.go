package lsp

import "encoding/json"

type Message struct {
	ID     int32           `json:"id"`
	Result json.RawMessage `json:"result"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
}

type Location struct {
	URI string `json:"uri"`
}

const (
	SymbolKindField = 8
)

type DocumentSymbol struct {
	Name           string           `json:"name"`
	Kind           int              `json:"kind"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children"`
	Detail         string           `json:"detail"`
}
