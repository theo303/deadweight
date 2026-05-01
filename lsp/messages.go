package lsp

import (
	"encoding/json"
	"fmt"
)

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
	End   Position `json:"end"`
}

type Location struct {
	URI string `json:"uri"`
}

type SymbolKind int

const (
	SymbolKindFile          SymbolKind = 1
	SymbolKindModule        SymbolKind = 2
	SymbolKindNamespace     SymbolKind = 3
	SymbolKindPackage       SymbolKind = 4
	SymbolKindClass         SymbolKind = 5
	SymbolKindMethod        SymbolKind = 6
	SymbolKindProperty      SymbolKind = 7
	SymbolKindField         SymbolKind = 8
	SymbolKindConstructor   SymbolKind = 9
	SymbolKindEnum          SymbolKind = 10
	SymbolKindInterface     SymbolKind = 11
	SymbolKindFunction      SymbolKind = 12
	SymbolKindVariable      SymbolKind = 13
	SymbolKindConstant      SymbolKind = 14
	SymbolKindString        SymbolKind = 15
	SymbolKindNumber        SymbolKind = 16
	SymbolKindBoolean       SymbolKind = 17
	SymbolKindArray         SymbolKind = 18
	SymbolKindObject        SymbolKind = 19
	SymbolKindKey           SymbolKind = 20
	SymbolKindNull          SymbolKind = 21
	SymbolKindEnumMember    SymbolKind = 22
	SymbolKindStruct        SymbolKind = 23
	SymbolKindEvent         SymbolKind = 24
	SymbolKindOperator      SymbolKind = 25
	SymbolKindTypeParameter SymbolKind = 26
)

var symbolKindNames = map[SymbolKind]string{
	SymbolKindFile:          "File",
	SymbolKindModule:        "Module",
	SymbolKindNamespace:     "Namespace",
	SymbolKindPackage:       "Package",
	SymbolKindClass:         "Class",
	SymbolKindMethod:        "Method",
	SymbolKindProperty:      "Property",
	SymbolKindField:         "Field",
	SymbolKindConstructor:   "Constructor",
	SymbolKindEnum:          "Enum",
	SymbolKindInterface:     "Interface",
	SymbolKindFunction:      "Function",
	SymbolKindVariable:      "Variable",
	SymbolKindConstant:      "Constant",
	SymbolKindString:        "String",
	SymbolKindNumber:        "Number",
	SymbolKindBoolean:       "Boolean",
	SymbolKindArray:         "Array",
	SymbolKindObject:        "Object",
	SymbolKindKey:           "Key",
	SymbolKindNull:          "Null",
	SymbolKindEnumMember:    "EnumMember",
	SymbolKindStruct:        "Struct",
	SymbolKindEvent:         "Event",
	SymbolKindOperator:      "Operator",
	SymbolKindTypeParameter: "TypeParameter",
}

func (sk SymbolKind) String() string {
	if name, ok := symbolKindNames[sk]; ok {
		return name
	}
	return fmt.Sprintf("Unknown (%d)", sk)
}

func ParseSymbolKind(s string) (SymbolKind, error) {
	for k, v := range symbolKindNames {
		if v == s {
			return k, nil
		}
	}
	return 0, fmt.Errorf("unknown symbol kind: %s", s)
}

type DocumentSymbol struct {
	Name           string           `json:"name"`
	Kind           SymbolKind       `json:"kind"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children"`
	Detail         string           `json:"detail"`
}
