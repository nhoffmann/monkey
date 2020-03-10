package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	store             map[string]Symbol
	numberDefinitions int
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func (st *SymbolTable) Define(symbolName string) Symbol {
	symbol := Symbol{Name: symbolName, Scope: GlobalScope, Index: st.numberDefinitions}
	st.store[symbolName] = symbol
	st.numberDefinitions++
	return symbol
}

func (st *SymbolTable) Resolve(symbolName string) (Symbol, bool) {
	symbol, ok := st.store[symbolName]
	return symbol, ok
}
