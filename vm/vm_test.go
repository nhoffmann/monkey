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
			{"1 < 2", true},
			{"1 > 2", false},
			{"1 < 1", false},
			{"1 > 1", false},
			{"1 == 1", true},
			{"1 == 2", false},
			{"1 != 2", true},
			{"-5", -5},
			{"-10", -10},
			{"-50 + 100 + -50", 0},
			{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
		}

		runVmTests(t, tests)
	})

	t.Run("Boolean expression", func(t *testing.T) {
		tests := []vmTestCase{
			{"true", true},
			{"false", false},
			{"true == true", true},
			{"false == false", true},
			{"true == false", false},
			{"true != false", true},
			{"false != true", true},
			{"true == true", true},
			{"(1 < 2) == true", true},
			{"(1 < 2) == false", false},
			{"(1 > 2) == true", false},
			{"(1 > 2) == false", true},
			{"!true", false},
			{"!false", true},
			{"!5", false},
			{"!!true", true},
			{"!!false", false},
			{"!!5", true},
			{"!if (false) { 10 }", true},
		}

		runVmTests(t, tests)
	})

	t.Run("Conditionals", func(t *testing.T) {
		tests := []vmTestCase{
			{"if (true) { 10 }", 10},
			{"if (true) { 10 } else { 20 }", 10},
			{"if (false) { 10 } else { 20 }", 20},
			{"if (1) { 10 }", 10},
			{"if (1 < 2) { 10 }", 10},
			{"if (1 < 2) { 10 } else { 20 }", 10},
			{"if (1 > 2) { 10 } else { 20 }", 20},
			{"if (1 > 2) { 10 }", Null},
			{"if (false) { 10 }", Null},
			{"if (if (false) { 10 }) { 10 } else { 20 }", 20},
		}

		runVmTests(t, tests)
	})

	t.Run("Global let statements", func(t *testing.T) {
		tests := []vmTestCase{
			{"let one = 1; one", 1},
			{"let one = 1; let two = 2; one + two;", 3},
			{"let one = 1; let two = one + one; one + two", 3},
		}

		runVmTests(t, tests)
	})

	t.Run("String expressions", func(t *testing.T) {
		tests := []vmTestCase{
			{`"monkey"`, "monkey"},
			{`"mon" + "key"`, "monkey"},
			{`"mon" + "key" + "banana"`, "monkeybanana"},
		}

		runVmTests(t, tests)
	})

	t.Run("Array literals", func(t *testing.T) {
		tests := []vmTestCase{
			{"[]", []int{}},
			{"[1, 2, 3]", []int{1, 2, 3}},
			{"[1 + 2, 3 * 4, 5 + 6]", []int{3, 12, 11}},
		}

		runVmTests(t, tests)
	})

	t.Run("Hash literals", func(t *testing.T) {
		tests := []vmTestCase{
			{"{}", map[object.HashKey]int64{}},
			{"{1: 2, 2: 3}", map[object.HashKey]int64{
				(&object.Integer{Value: 1}).HashKey(): 2,
				(&object.Integer{Value: 2}).HashKey(): 3,
			}},
			{"{1 + 1: 2 * 2, 3 + 3: 4 * 4}", map[object.HashKey]int64{
				(&object.Integer{Value: 2}).HashKey(): 4,
				(&object.Integer{Value: 6}).HashKey(): 16,
			}},
		}

		runVmTests(t, tests)
	})

	t.Run("", func(t *testing.T) {
		tests := []vmTestCase{
			{"[1, 2, 3][1]", 2},
			{"[1, 2, 3][0 + 2]", 3},
			{"[[1, 2, 3]][0][0]", 1},
			{"[][0]", Null},
			{"[1, 2, 3][99]", Null},
			{"[1][-1]", Null},
			{"{1: 1, 2: 2}[1]", 1},
			{"{1: 1, 2: 2}[2]", 2},
			{"{1: 1}[0]", Null},
			{"{}[0]", Null},
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
	case []int:
		assertIntegerArray(t, actual, expected)
	case bool:
		assertBooleanObject(t, actual, bool(expected))
	case string:
		assertStringObject(t, actual, expected)
	case map[object.HashKey]int64:
		assertHashWithIntegerKeysAndValues(t, actual, expected)
	case *object.Null:
		assertNull(t, actual)
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

func assertIntegerArray(t *testing.T, actual object.Object, expected []int) {
	array, ok := actual.(*object.Array)
	if !ok {
		t.Errorf("Object is not an array. Got %T: %+v", actual, actual)
	}

	if len(array.Elements) != len(expected) {
		t.Errorf("Wrong number of elements. Got %d, want %d", len(array.Elements), len(expected))
	}

	for i, expectedElement := range expected {
		assertIntegerObject(t, array.Elements[i], int64(expectedElement))
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

func assertNull(t *testing.T, evaluated object.Object) {
	t.Helper()

	if evaluated != Null {
		t.Errorf("Object is not Null: %T %+v", evaluated, evaluated)
	}
}

func assertStringObject(t *testing.T, actual object.Object, expected string) {
	t.Helper()

	stringObject, ok := actual.(*object.String)

	if !ok {
		t.Errorf("Object is not a string. Got %T: %+v", actual, actual)
	} else {
		if stringObject.Value != expected {
			t.Errorf("Object has improper value. Expected %s, got %s", expected, stringObject.Value)
		}
	}
}

func assertHashWithIntegerKeysAndValues(t *testing.T, actual object.Object, expected map[object.HashKey]int64) {
	t.Helper()

	hash, ok := actual.(*object.Hash)
	if !ok {
		t.Errorf("Object is not a hash. Got %T: %+v", actual, actual)
	}

	if len(hash.Pairs) != len(expected) {
		t.Errorf("Hash has wrong number of pairs. Expected %d, got %d.", len(expected), len(hash.Pairs))
	}

	for expectedKey, expectedValue := range expected {
		pair, ok := hash.Pairs[expectedKey]
		if !ok {
			t.Errorf("No pair for a given key in Pairs. Key: %q", expectedKey)
		}

		assertIntegerObject(t, pair.Value, expectedValue)
	}
}
