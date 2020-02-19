package evaluator

import (
	"testing"

	"github.com/nhoffmann/monkey/lexer"
	"github.com/nhoffmann/monkey/object"
	"github.com/nhoffmann/monkey/parser"
)

func TestEvaluator(t *testing.T) {
	t.Run("Evaluate Integer Expression", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"5", 5},
			{"10", 10},
			{"-5", -5},
			{"-10", -10},
		}

		for _, test := range tests {
			evaluated := evaluateInput(t, test.input)

			assertIntegerObject(t, evaluated, test.expected)
		}
	})

	t.Run("Evaluate Boolean Expression", func(t *testing.T) {
		tests := []struct {
			input    string
			expected bool
		}{
			{"true", true},
			{"false", false},
		}

		for _, test := range tests {
			evaluated := evaluateInput(t, test.input)

			assertBooleanObject(t, evaluated, test.expected)
		}
	})

	t.Run("Evaluate Bang Operator", func(t *testing.T) {
		tests := []struct {
			input    string
			expected bool
		}{
			{"!true", false},
			{"!false", true},
			{"!5", false},
			{"!!true", true},
			{"!!false", false},
			{"!!5", true},
		}

		for _, test := range tests {
			evaluated := evaluateInput(t, test.input)

			assertBooleanObject(t, evaluated, test.expected)
		}
	})
}

func assertIntegerObject(t *testing.T, evaluated object.Object, want int64) {
	t.Helper()

	integerObject, ok := evaluated.(*object.Integer)

	if !ok {
		t.Fatalf("Object is not an integer. Got %T: %+v", evaluated, evaluated)
	}

	if integerObject.Value != want {
		t.Errorf("Object has improper value. Expected %d, got %d", want, integerObject.Value)
	}
}

func assertBooleanObject(t *testing.T, evaluated object.Object, want bool) {
	t.Helper()

	booleanObject, ok := evaluated.(*object.Boolean)

	if !ok {
		t.Fatalf("Object is not a boolean. Got %t: %+v", evaluated, evaluated)
	}

	if booleanObject.Value != want {
		t.Errorf("Object has improper value. Expected %t, got %t", want, booleanObject.Value)
	}
}

func evaluateInput(t *testing.T, input string) object.Object {
	t.Helper()

	lexer := lexer.NewLexer(input)
	parser := parser.NewParser(lexer)

	program := parser.ParseProgram()

	return Eval(program)
}
