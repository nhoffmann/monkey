package compiler

import (
	"testing"

	"github.com/nhoffmann/monkey/object"

	"github.com/nhoffmann/monkey/ast"
	"github.com/nhoffmann/monkey/code"
	"github.com/nhoffmann/monkey/lexer"
	"github.com/nhoffmann/monkey/parser"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []code.Instructions
}

func TestCompiler(t *testing.T) {
	t.Run("Integer arithmetic", func(t *testing.T) {
		tests := []compilerTestCase{
			{
				input:             "1 + 2",
				expectedConstants: []interface{}{1, 2},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpPop),
				},
			},
			{
				input:             "1; 2;",
				expectedConstants: []interface{}{1, 2},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpPop),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpPop),
				},
			},
			{
				input:             "1 - 2",
				expectedConstants: []interface{}{1, 2},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSubtract),
					code.Make(code.OpPop),
				},
			},
			{
				input:             "1 * 2",
				expectedConstants: []interface{}{1, 2},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpMultiply),
					code.Make(code.OpPop),
				},
			},
			{
				input:             "2 / 1",
				expectedConstants: []interface{}{2, 1},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpDivide),
					code.Make(code.OpPop),
				},
			},
			{
				input:             "-1",
				expectedConstants: []interface{}{1},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpMinus),
					code.Make(code.OpPop),
				},
			},
		}

		runCompilerTests(t, tests)
	})

	t.Run("Boolean expressions", func(t *testing.T) {
		tests := []compilerTestCase{
			{
				input:             "true",
				expectedConstants: []interface{}{},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpTrue),
					code.Make(code.OpPop),
				},
			},
			{
				input:             "false",
				expectedConstants: []interface{}{},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpFalse),
					code.Make(code.OpPop),
				},
			},
			{
				input:             "1 > 2",
				expectedConstants: []interface{}{1, 2},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpGreaterThan),
					code.Make(code.OpPop),
				},
			},
			{
				input:             "1 < 2",
				expectedConstants: []interface{}{2, 1},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpGreaterThan),
					code.Make(code.OpPop),
				},
			},
			{
				input:             "1 == 2",
				expectedConstants: []interface{}{1, 2},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpEqual),
					code.Make(code.OpPop),
				},
			},
			{
				input:             "1 != 2",
				expectedConstants: []interface{}{1, 2},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpNotEqual),
					code.Make(code.OpPop),
				},
			},
			{
				input:             "true == false",
				expectedConstants: []interface{}{},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpTrue),
					code.Make(code.OpFalse),
					code.Make(code.OpEqual),
					code.Make(code.OpPop),
				},
			},
			{
				input:             "true != false",
				expectedConstants: []interface{}{},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpTrue),
					code.Make(code.OpFalse),
					code.Make(code.OpNotEqual),
					code.Make(code.OpPop),
				},
			},
			{
				input:             "!true",
				expectedConstants: []interface{}{},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpTrue),
					code.Make(code.OpBang),
					code.Make(code.OpPop),
				},
			},
		}

		runCompilerTests(t, tests)
	})

	t.Run("Conditionals", func(t *testing.T) {
		tests := []compilerTestCase{
			{
				input:             `if (true) { 10 }; 3333;`,
				expectedConstants: []interface{}{10, 3333},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpTrue),              // 0000
					code.Make(code.OpJumpNotTruthy, 10), // 0001
					code.Make(code.OpConstant, 0),       // 0004
					code.Make(code.OpJump, 11),          // 0007
					code.Make(code.OpNull),              // 0010
					code.Make(code.OpPop),               // 0011
					code.Make(code.OpConstant, 1),       // 0012
					code.Make(code.OpPop),               // 0015
				},
			},
			{
				input:             `if (true) { 10 } else { 20 }; 3333;`,
				expectedConstants: []interface{}{10, 20, 3333},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpTrue),              // 0000
					code.Make(code.OpJumpNotTruthy, 10), // 0001
					code.Make(code.OpConstant, 0),       // 0004
					code.Make(code.OpJump, 13),          // 0007
					code.Make(code.OpConstant, 1),       // 0010
					code.Make(code.OpPop),               // 0013
					code.Make(code.OpConstant, 2),       // 0014
					code.Make(code.OpPop),               // 0017
				},
			},
		}

		runCompilerTests(t, tests)
	})

	t.Run("Global let statements", func(t *testing.T) {
		tests := []compilerTestCase{
			{
				input: `
					let one = 1;
					let two = 2;
				`,
				expectedConstants: []interface{}{1, 2},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetGlobal, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSetGlobal, 1),
				},
			},
			{
				input: `
					let one = 1;
					one;
				`,
				expectedConstants: []interface{}{1},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetGlobal, 0),
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpPop),
				},
			},
			{
				input: `
					let one = 1;
					let two = one;
					two;
				`,
				expectedConstants: []interface{}{1},
				expectedInstructions: []code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetGlobal, 0),
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpSetGlobal, 1),
					code.Make(code.OpGetGlobal, 1),
					code.Make(code.OpPop),
				},
			},
		}

		runCompilerTests(t, tests)
	})
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, test := range tests {
		program := parse(test.input)

		compiler := NewCompiler()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		assertInstructions(t, test.expectedInstructions, bytecode.Instructions)
		assertConstants(t, test.expectedConstants, bytecode.Constants)
	}
}

func parse(input string) *ast.Program {
	lexer := lexer.NewLexer(input)
	parser := parser.NewParser(lexer)

	return parser.ParseProgram()
}

func assertInstructions(t *testing.T, expected []code.Instructions, actual code.Instructions) {
	t.Helper()

	concattedExpected := concatInstructions(expected)

	if len(actual) != len(concattedExpected) {
		t.Fatalf("Instructions: wrong length.\nWant %q,\ngot  %q", concattedExpected, actual)
	}

	for i, expectedInstruction := range concattedExpected {
		if actual[i] != expectedInstruction {
			t.Fatalf(
				"Instructions: wrong instruction at %d.\nWant %q,\ngot  %q",
				i,
				concattedExpected,
				actual,
			)
		}
	}
}

func concatInstructions(instructions []code.Instructions) code.Instructions {
	out := code.Instructions{}

	for _, instruction := range instructions {
		out = append(out, instruction...)
	}

	return out
}

func assertConstants(t *testing.T, expected []interface{}, actual []object.Object) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Fatalf("Constants: wrong number of constants. Want %d, got %d", len(expected), len(actual))
	}

	for i, expectedConstant := range expected {
		switch expectedConstant := expectedConstant.(type) {
		case int:
			assertIntegerObject(t, actual[i], int64(expectedConstant))
		}
	}
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
