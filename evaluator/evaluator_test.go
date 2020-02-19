package evaluator

import (
	"testing"

	"github.com/nhoffmann/monkey/lexer"
	"github.com/nhoffmann/monkey/object"
	"github.com/nhoffmann/monkey/parser"
)

func TestEvaluator(t *testing.T) {
	t.Run("Integer Expression", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"5", 5},
			{"10", 10},
		}

		for _, test := range tests {
			evaluated := evaluateInput(t, test.input)

			assertIntegerObject(t, evaluated, test.expected)
		}
	})
}

func assertIntegerObject(t *testing.T, evaluated object.Object, want int64) {
	t.Helper()

	integerObject, ok := evaluated.(*object.Integer)

	if !ok {
		t.Fatalf("Object is an integer. Got %T: %+v", evaluated, evaluated)
	}

	if integerObject.Value != want {
		t.Errorf("Object has improper value. Expected %d, got %d", want, integerObject.Value)
	}
}

func evaluateInput(t *testing.T, input string) object.Object {
	t.Helper()

	lexer := lexer.NewLexer(input)
	parser := parser.NewParser(lexer)

	program := parser.ParseProgram()

	return Eval(program)
}
