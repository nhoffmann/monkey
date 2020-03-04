package vm

import (
	"testing"

	"github.com/nhoffmann/monkey/compiler"

	"github.com/nhoffmann/monkey/ast"
	"github.com/nhoffmann/monkey/lexer"
	"github.com/nhoffmann/monkey/object"
	"github.com/nhoffmann/monkey/parser"
)

type vmTestCase struct {
	input    string
	expected interface{}
}

func TestCases(t *testing.T) {
	t.Run("Integer arithmetic", func(t *testing.T) {
		tests := []vmTestCase{
			{"1", 1},
			{"2", 2},
			{"1 + 2", 3},
			{"1 - 2", -1},
			{"1 * 2", 2},
			{"4 / 2", 2},
			{"50 / 2 * 2 + 10 - 5", 55},
			{"5 * (2 + 10)", 60},
			{"5 + 5 + 5 + 5 - 10", 10},
			{"2 * 2 * 2 * 2 * 2", 32},
			{"5 * 2 + 10", 20},
			{"5 + 2 * 10", 25},
		}

		runVmTests(t, tests)
	})

	t.Run("Boolean expression", func(t *testing.T) {
		tests := []vmTestCase{
			{"true", true},
			{"false", false},
		}

		runVmTests(t, tests)
	})
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, test := range tests {
		program := parse(test.input)

		compiler := compiler.NewCompiler()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := NewVm(compiler.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElement := vm.LastPoppedStackElement()

		assertExpectedObject(t, stackElement, test.expected)
	}
}

func assertExpectedObject(t *testing.T, actual object.Object, expected interface{}) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		assertIntegerObject(t, actual, int64(expected))
	case bool:
		assertBooleanObject(t, actual, bool(expected))
	}
}

func parse(input string) *ast.Program {
	lexer := lexer.NewLexer(input)
	parser := parser.NewParser(lexer)

	return parser.ParseProgram()
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
