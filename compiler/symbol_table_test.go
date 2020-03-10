package compiler

import "testing"

func TestSymbolTable(t *testing.T) {
	t.Run("Define", func(t *testing.T) {
		expected := map[string]Symbol{
			"a": Symbol{Name: "a", Scope: GlobalScope, Index: 0},
			"b": Symbol{Name: "b", Scope: GlobalScope, Index: 1},
		}

		global := NewSymbolTable()

		a := global.Define("a")
		if a != expected["a"] {
			t.Errorf("Expected a=%+v, got %+v", expected["a"], a)
		}

		b := global.Define("b")
		if b != expected["b"] {
			t.Errorf("Expected b=%+v, got %+v", expected["b"], b)
		}
	})

	t.Run("Resolve global", func(t *testing.T) {
		global := NewSymbolTable()
		global.Define("a")
		global.Define("b")

		expected := []Symbol{
			Symbol{Name: "a", Scope: GlobalScope, Index: 0},
			Symbol{Name: "b", Scope: GlobalScope, Index: 1},
		}

		for _, symbol := range expected {
			result, ok := global.Resolve(symbol.Name)

			if !ok {
				t.Errorf("Could not resolve name: %s", symbol.Name)
				continue
			}

			if result != symbol {
				t.Errorf(
					"Expected %s to resolve to %+v, got %+v",
					symbol.Name,
					symbol,
					result,
				)
			}
		}
	})
}
