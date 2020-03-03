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

	t.Run("Evaluate String literals", func(t *testing.T) {
		input := `"Hello World!"`

		evaluated := evaluateInput(t, input)

		str, ok := evaluated.(*object.String)
		if !ok {
			t.Fatalf("Object is not a string. Got %t: %+v", evaluated, evaluated)
		}

		if str.Value != "Hello World!" {
			t.Errorf("String has wrong value. Got %q", str.Value)
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
			{
				`
				let f = fn(x) {
				return x;
				x + 10;
				};
				f(10);`,
				10,
			},
			{
				`
				let f = fn(x) {
				let result = x + 10;
				return result;
				return 10;
				};
				f(10);`,
				20,
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
			{
				"foobar",
				"identifier not found: foobar",
			},
			{
				`"Hello" - "World"`,
				"unknown operator: STRING - STRING",
			},
		}

		for _, test := range tests {
			evaluated := evaluateInput(t, test.input)

			errorObject, ok := evaluated.(*object.Error)

			if !ok {
				t.Errorf("Expected object.Error, got %T: %+v", evaluated, evaluated)
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

	t.Run("Let Statements", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"let a = 5; a", 5},
			{"let a = 5 * 5; a", 25},
			{"let a = 5; let b = a; b", 5},
			{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
		}

		for _, test := range tests {
			evaluated := evaluateInput(t, test.input)

			assertIntegerObject(t, evaluated, test.expected)
		}
	})

	t.Run("Evaluate Function Object", func(t *testing.T) {
		input := "fn(x) { x + 2; };"

		evaluated := evaluateInput(t, input)

		function, ok := evaluated.(*object.Function)
		if !ok {
			t.Fatalf("Expected object.Function, got %T: %+v", evaluated, evaluated)
		}

		if len(function.Parameters) != 1 {
			t.Fatalf("Function has wrong parameters: %+v", function.Parameters)
		}

		if function.Parameters[0].String() != "x" {
			t.Fatalf("Parameter is not 'x'. Got %q", function.Parameters[0])
		}

		expectedBody := "(x + 2)"

		if function.Body.String() != expectedBody {
			t.Fatalf("Expected body to be %q, got %q", expectedBody, function.Body.String())
		}
	})

	t.Run("Function Application", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"let identity = fn(x) { x; }; identity(5);", 5},
			{"let identity = fn(x) { return x; }; identity(5);", 5},
			{"let double = fn(x) { x * 2; }; double(5);", 10},
			{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
			{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
			{"fn(x) { x; }(5)", 5},
		}

		for _, test := range tests {
			evaluated := evaluateInput(t, test.input)
			assertIntegerObject(t, evaluated, test.expected)
		}
	})

	t.Run("Closures", func(t *testing.T) {
		input := `
		let newAdder = fn(x) {
			fn(y) { x + y };
		};

		let addTwo = newAdder(2);
		addTwo(2)`

		assertIntegerObject(t, evaluateInput(t, input), 4)
	})

	t.Run("Builtin Function", func(t *testing.T) {
		tests := []struct {
			input    string
			expected interface{}
		}{
			{`len("")`, 0},
			{`len("four")`, 4},
			{`len("hello world")`, 11},
			{`len(1)`, "argument to `len` not supported, got INTEGER."},
			{`len("one", "two")`, "wrong number of arguments. Got 2, want 1."},
			{"let a = [1, 2, 3]; len(a)", 3},
		}

		for _, test := range tests {
			evaluated := evaluateInput(t, test.input)

			switch expected := test.expected.(type) {
			case int:
				assertIntegerObject(t, evaluated, int64(expected))
			case string:
				errorObject, ok := evaluated.(*object.Error)

				if !ok {
					t.Errorf("Object is not Error. Got %T: %+v", evaluated, evaluated)
					continue
				}

				if errorObject.Message != expected {
					t.Errorf("wrong error message. Expected %q, got %q", expected, errorObject.Message)
				}
			}
		}
	})

	t.Run("Array literals", func(t *testing.T) {
		input := "[1, 2 * 2, 3 + 3]"

		evaluated := evaluateInput(t, input)

		result, ok := evaluated.(*object.Array)
		if !ok {
			t.Fatalf("Object is not Array. Got %T: %+v", evaluated, evaluated)
		}

		if len(result.Elements) != 3 {
			t.Fatalf("Expected exactly 3 elements but got %d", len(result.Elements))
		}

		assertIntegerObject(t, result.Elements[0], 1)
		assertIntegerObject(t, result.Elements[1], 4)
		assertIntegerObject(t, result.Elements[2], 6)
	})

	t.Run("Array index expressions", func(t *testing.T) {
		tests := []struct {
			input    string
			expected interface{}
		}{
			{"[1, 2, 3][0]", 1},
			{"[1, 2, 3][1]", 2},
			{"[1, 2, 3][2]", 3},
			{"let i = 0; [1][i]", 1},
			{"[1, 2, 3][1 + 1]", 3},
			{"let a = [1, 2, 3]; a[2]", 3},
			{"let a = [1, 2, 3]; a[0] + a[1] + a[2]", 6},
			{"let a = [1, 2, 3]; let i = a[0]; a[i]", 2},
			{"[1, 2, 3][3]", nil},
			{"[1, 2, 3][-1]", nil},
		}

		for _, test := range tests {
			evaluated := evaluateInput(t, test.input)

			integer, ok := test.expected.(int)
			if ok {
				assertIntegerObject(t, evaluated, int64(integer))
			} else {
				assertNullObject(t, evaluated)
			}
		}
	})

	t.Run("Hash literals", func(t *testing.T) {
		input := `let two = "two";
		{
			"one": 10 - 9,
			two: 1 + 1,
			"thr" + "ee": 6 / 2,
			4: 4,
			true: 5,
			false: 6,
		}`

		evaluated := evaluateInput(t, input)

		hash, ok := evaluated.(*object.Hash)
		if !ok {
			t.Fatalf("Eval didn't return Hash. Got %T: %+v", evaluated, evaluated)
		}

		expected := map[object.HashKey]int64{
			(&object.String{Value: "one"}).HashKey():   1,
			(&object.String{Value: "two"}).HashKey():   2,
			(&object.String{Value: "three"}).HashKey(): 3,
			(&object.Integer{Value: 4}).HashKey():      4,
			TRUE.HashKey():                             5,
			FALSE.HashKey():                            6,
		}

		if len(hash.Pairs) != len(expected) {
			t.Fatalf(
				"Hash has wrong number of arguments. Want %d, got %d",
				len(hash.Pairs),
				len(expected),
			)
		}

		for expectedKey, expectedValue := range expected {
			pair, ok := hash.Pairs[expectedKey]
			if !ok {
				t.Errorf("No pair for given key in Pairs")
			}

			assertIntegerObject(t, pair.Value, expectedValue)
		}
	})

	t.Run("Hash index expressions", func(t *testing.T) {
		tests := []struct {
			input    string
			expected interface{}
		}{
			{
				`{"foo": 5}["foo"]`,
				5,
			},
			{
				`{"foo": 5}["bar"]`,
				nil,
			},
			{
				`let key = "foo"; {"foo": 5}[key]`,
				5,
			},
			{
				`{}["foo"]`,
				nil,
			},
			{
				`{"foo": 5}["foo"]`,
				5,
			},
			{
				`{5: 5}[5]`,
				5,
			},
			{
				`{true: 5}[true]`,
				5,
			},
			{
				`{false: 5}[false]`,
				5,
			},
		}

		for _, test := range tests {
			evaluated := evaluateInput(t, test.input)

			integer, ok := test.expected.(int)

			if ok {
				assertIntegerObject(t, evaluated, int64(integer))
			} else {
				assertNullObject(t, evaluated)
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
	env := object.NewEnvironment()

	return Eval(program, env)
}
