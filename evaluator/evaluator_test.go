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
			{"5 + 5 + 5 + 5 - 10", 10},
			{"2 * 2 * 2 * 2 * 2", 32},
			{"-50 + 100 + -50", 0},
			{"5 * 2 + 10", 20},
			{"5 + 2 * 10", 25},
			{"20 + 2 * -10", 0},
			{"50 / 2 * 2 + 10", 60},
			{"2 * (5 + 10)", 30},
			{"3 * 3 * 3 + 10", 37},
			{"3 * (3 * 3) + 10", 37},
			{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
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
			{"1 < 2", true},
			{"1 > 2", false},
			{"1 < 1", false},
			{"1 > 1", false},
			{"1 == 1", true},
			{"1 != 1", false},
			{"1 == 2", false},
			{"1 != 2", true},
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
			{"true == true", true},
			{"false == false", true},
			{"true == false", false},
			{"true != false", true},
			{"(1 < 2) == true", true},
			{"(1 < 2) == false", false},
			{"(1 > 2) == true", false},
			{"(1 > 2) == false", true},
		}

		for _, test := range tests {
			evaluated := evaluateInput(t, test.input)

			assertBooleanObject(t, evaluated, test.expected)
		}
	})

	t.Run("Evaluate If Else Expression", func(t *testing.T) {
		tests := []struct {
			input    string
			expected interface{}
		}{
			{"if (true) { 10 }", 10},
			{"if (false) { 10 }", nil},
			// {"if (1) { 10 }", 10},
			{"if (1 < 2) { 10 }", 10},
			{"if (1 > 2) { 10 }", nil},
			{"if (1 > 2) { 10 } else { 20 }", 20},
			{"if (1 < 2) { 10 } else { 20 }", 10},
		}

		for _, test := range tests {
			evaluated := evaluateInput(t, test.input)

			integerObject, ok := test.expected.(int)

			if ok {
				assertIntegerObject(t, evaluated, int64(integerObject))
			} else {
				assertNullObject(t, evaluated)
			}
		}
	})

	t.Run("Evaluate Return Statements", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"return 10", 10},
			{"return 10; 9;", 10},
			{"return 2 * 5; 9", 10},
			{"9; return 2 * 5; 9", 10},
			{
				`if(10 > 1) {
					if(10 > 1) {
						return 10;
					}

					return 1;
				}
				`,
				10,
			},
		}

		for _, test := range tests {
			evaluated := evaluateInput(t, test.input)

			assertIntegerObject(t, evaluated, test.expected)
		}
	})

	t.Run("Error handling", func(t *testing.T) {
		tests := []struct {
			input           string
			expectedMessage string
		}{
			{
				"5 + true",
				"type mismatch: INTEGER + BOOLEAN",
			},
			{
				"5 + true; 5",
				"type mismatch: INTEGER + BOOLEAN",
			},
			{
				"-true",
				"unknown operator: -BOOLEAN",
			},
			{
				"true + false",
				"unknown operator: BOOLEAN + BOOLEAN",
			},
			{
				"5; true + false; 5",
				"unknown operator: BOOLEAN + BOOLEAN",
			},
			{
				"if (10 > 1) { true + false }",
				"unknown operator: BOOLEAN + BOOLEAN",
			},
			{
				`if(10 > 1) {
					if(10 > 1) {
						return true + false;
					}

					return 1;
				}
				`,
				"unknown operator: BOOLEAN + BOOLEAN",
			},
		}

		for _, test := range tests {
			evaluated := evaluateInput(t, test.input)

			errorObject, ok := evaluated.(*object.Error)

			if !ok {
				t.Errorf("Expected an error, got %T: %+v", evaluated, evaluated)
				continue
			}

			if errorObject.Message != test.expectedMessage {
				t.Errorf(
					"Wrong error message. Expected %q, got %q",
					test.expectedMessage,
					errorObject.Message,
				)
			}
		}
	})
}

func assertIntegerObject(t *testing.T, evaluated object.Object, want int64) {
	t.Helper()

	integerObject, ok := evaluated.(*object.Integer)

	if !ok {
		t.Errorf("Object is not an integer. Got %T: %+v", evaluated, evaluated)
	} else {
		if integerObject.Value != want {
			t.Errorf("Object has improper value. Expected %d, got %d", want, integerObject.Value)
		}
	}

}

func assertBooleanObject(t *testing.T, evaluated object.Object, want bool) {
	t.Helper()

	booleanObject, ok := evaluated.(*object.Boolean)

	if !ok {
		t.Errorf("Object is not a boolean. Got %t: %+v", evaluated, evaluated)
	} else {
		if booleanObject.Value != want {
			t.Errorf("Object has improper value. Expected %t, got %t", want, booleanObject.Value)
		}
	}

}

func assertNullObject(t *testing.T, evaluated object.Object) {
	if evaluated != NULL {
		t.Errorf("Object is not NULL. Got %T:%+v", evaluated, evaluated)
	}
}

func evaluateInput(t *testing.T, input string) object.Object {
	t.Helper()

	lexer := lexer.NewLexer(input)
	parser := parser.NewParser(lexer)

	program := parser.ParseProgram()

	return Eval(program)
}
