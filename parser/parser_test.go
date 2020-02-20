package parser

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/nhoffmann/monkey/ast"
	"github.com/nhoffmann/monkey/lexer"
	"github.com/nhoffmann/monkey/token"
)

func TestParseProgram(t *testing.T) {
	t.Run("Parser errors", func(t *testing.T) {
		t.Run("PeekError", func(t *testing.T) {
			tests := []struct {
				input             string
				expectedTokenType token.TokenType
				actualTokenType   token.TokenType
			}{
				{"let x 5;", token.ASSIGN, token.INT},
				{"let = 10;", token.IDENT, token.ASSIGN},
				{"let 838383;", token.IDENT, token.INT},
			}

			for _, test := range tests {
				lexer := lexer.NewLexer(test.input)
				parser := NewParser(lexer)

				parser.ParseProgram()

				if len(parser.Errors()) == 0 {
					t.Error("Expected errors to be present")
				}

				error, ok := parser.Errors()[0].(*PeekError)

				if !ok {
					t.Fatal("Expected PeekError but got", error)
				}

				if error.Error() != fmt.Sprintf("Expected next token to be %q, but got %q", test.expectedTokenType, test.actualTokenType) {
					t.Error(error)
				}
			}
		})
	})

	t.Run("Parse let statements", func(t *testing.T) {
		tests := []struct {
			input              string
			expectedIdentifier string
			expectedValue      interface{}
		}{
			{"let x = 5;", "x", 5},
			{"let y = true;", "y", true},
			{"let foobar = y;", "foobar", "y"},
		}

		for _, test := range tests {
			program := parseInput(t, test.input)

			assertStatementsPresent(t, program)

			letStatement, ok := program.Statements[0].(*ast.LetStatement)

			assertNodeType(t, ok, letStatement, "*ast.LetStatement")
			assertTokenLiteral(t, letStatement, "let")

			assertIdentifierValue(t, letStatement.Name, test.expectedIdentifier)
			assertIdentifierTokenLiteral(t, letStatement.Name, test.expectedIdentifier)
		}
	})

	t.Run("Parse return statement", func(t *testing.T) {
		tests := []struct {
			input string
		}{
			{"return 5;"},
			{"return 10;"},
			{"return 993322;"},
		}

		for _, test := range tests {
			program := parseInput(t, test.input)

			assertStatementsPresent(t, program)

			returnStatement, ok := program.Statements[0].(*ast.ReturnStatement)

			assertNodeType(t, ok, returnStatement, "*ast.ReturnStatement")
			assertTokenLiteral(t, returnStatement, "return")
		}
	})

	t.Run("Parse Identifier expression", func(t *testing.T) {
		input := "foobar;"

		program := parseInput(t, input)

		assertStatementsPresent(t, program)

		expressionStatement, ok := program.Statements[0].(*ast.ExpressionStatement)

		assertNodeType(t, ok, expressionStatement, "*ast.ExpressionStatement")

		ident, ok := expressionStatement.Expression.(*ast.Identifier)

		assertNodeType(t, ok, ident, "*ast.Identifier")
		assertIdentifierValue(t, ident, "foobar")
		assertTokenLiteral(t, ident, "foobar")
	})

	t.Run("Parse Integer expression", func(t *testing.T) {
		input := "5;"

		program := parseInput(t, input)

		assertStatementsPresent(t, program)

		expressionStatement, ok := program.Statements[0].(*ast.ExpressionStatement)

		assertNodeType(t, ok, expressionStatement, "*ast.ExpressionStatement")
		assertIntegerLiteral(t, expressionStatement.Expression, 5)
	})

	t.Run("Parse boolean expression", func(t *testing.T) {
		tests := []struct {
			input           string
			expectedBoolean bool
		}{
			{"true;", true},
			{"false;", false},
		}

		for _, test := range tests {
			program := parseInput(t, test.input)

			assertStatementsPresent(t, program)

			expressionStatement, ok := program.Statements[0].(*ast.ExpressionStatement)
			assertNodeType(t, ok, expressionStatement, "*ast.ExpressionStatement")

			assertBooleanLiteral(t, expressionStatement.Expression, test.expectedBoolean)
		}
	})

	t.Run("Parse string literal", func(t *testing.T) {
		input := `"hello world"`

		program := parseInput(t, input)

		assertStatementsPresent(t, program)

		expressionStatement, ok := program.Statements[0].(*ast.ExpressionStatement)
		stringLiteral, ok := expressionStatement.Expression.(*ast.StringLiteral)
		assertNodeType(t, ok, stringLiteral, "*ast.StringLiteral")

		expectedValue := "hello world"

		if stringLiteral.Value != expectedValue {
			t.Errorf(
				"String literal value not correct. Expected %q, got %q",
				expectedValue,
				stringLiteral.Value,
			)
		}
	})

	t.Run("Parse prefix expression", func(t *testing.T) {
		tests := []struct {
			input    string
			operator string
			value    interface{}
		}{
			{"!5", "!", 5},
			{"-15", "-", 15},
			{"!true", "!", true},
			{"!false", "!", false},
		}

		for _, test := range tests {
			program := parseInput(t, test.input)

			assertStatementsPresent(t, program)

			expressionStatement, ok := program.Statements[0].(*ast.ExpressionStatement)

			assertNodeType(t, ok, expressionStatement, "*ast.ExpressionStatement")

			prefix, ok := expressionStatement.Expression.(*ast.PrefixExpression)

			assertNodeType(t, ok, prefix, "*ast.PrefixExpression")
			assertOperator(t, prefix.Operator, test.operator)

			assertLiteralExpression(t, prefix.Right, test.value)
		}
	})

	t.Run("Parse infix expressions", func(t *testing.T) {
		tests := []struct {
			input      string
			leftValue  interface{}
			operator   string
			rightValue interface{}
		}{
			{"5 + 5;", 5, "+", 5},
			{"5 - 5;", 5, "-", 5},
			{"5 * 5;", 5, "*", 5},
			{"5 / 5;", 5, "/", 5},
			{"5 > 5;", 5, ">", 5},
			{"5 < 5;", 5, "<", 5},
			{"5 == 5;", 5, "==", 5},
			{"5 != 5;", 5, "!=", 5},
			{"true == true", true, "==", true},
			{"true != false", true, "!=", false},
			{"false == false", false, "==", false},
		}

		for _, test := range tests {
			program := parseInput(t, test.input)

			assertStatementsPresent(t, program)

			expressionStatement, ok := program.Statements[0].(*ast.ExpressionStatement)

			assertNodeType(t, ok, expressionStatement, "*ast.ExpressionStatement")

			infix, ok := expressionStatement.Expression.(*ast.InfixExpression)

			assertInfixExpression(t, infix, test.leftValue, test.operator, test.rightValue)
		}
	})

	t.Run("If Expression", func(t *testing.T) {
		input := `if (x < y) { x }`

		program := parseInput(t, input)

		assertStatementsPresent(t, program)
		expressionStatement, ok := program.Statements[0].(*ast.ExpressionStatement)

		assertNodeType(t, ok, expressionStatement, "*ast.ExpressionStatement")

		ifExpression, ok := expressionStatement.Expression.(*ast.IfExpression)

		assertInfixExpression(t, ifExpression.Condition, "x", "<", "y")

		assertBlockExpression(t, ifExpression.Consequence, "x")

		if ifExpression.Alternative != nil {
			t.Errorf("Did not expect alternative but got one: %+v", ifExpression.Alternative)
		}
	})

	t.Run("If Else Expression", func(t *testing.T) {
		input := `if (x < y) { x } else { y }`

		program := parseInput(t, input)
		assertStatementsPresent(t, program)

		expressionStatement, ok := program.Statements[0].(*ast.ExpressionStatement)
		assertNodeType(t, ok, expressionStatement, "*ast.ExpressionStatement")

		ifExpression, ok := expressionStatement.Expression.(*ast.IfExpression)

		assertInfixExpression(t, ifExpression.Condition, "x", "<", "y")

		assertBlockExpression(t, ifExpression.Consequence, "x")
		assertBlockExpression(t, ifExpression.Alternative, "y")
	})

	t.Run("Function Literal", func(t *testing.T) {
		input := `fn(x, y) { x + y; }`

		program := parseInput(t, input)
		assertStatementsPresent(t, program)

		expressionStatement, ok := program.Statements[0].(*ast.ExpressionStatement)
		assertNodeType(t, ok, expressionStatement, "*ast.ExpressionStatement")

		functionLiteral, ok := expressionStatement.Expression.(*ast.FunctionLiteral)
		assertNodeType(t, ok, functionLiteral, "*ast.FunctionLiteral")

		if len(functionLiteral.Parameters) != 2 {
			t.Errorf("Expected 2 parameters, got %d", len(functionLiteral.Parameters))
		}

		assertLiteralExpression(t, functionLiteral.Parameters[0], "x")
		assertLiteralExpression(t, functionLiteral.Parameters[1], "y")

		assertBlockExpression(t, functionLiteral.Body, nil)

		bodyStatement, ok := functionLiteral.Body.Statements[0].(*ast.ExpressionStatement)
		assertNodeType(t, ok, functionLiteral, "*ast.ExpressionStatement")

		assertInfixExpression(t, bodyStatement.Expression, "x", "+", "y")
	})

	t.Run("Parse function parameters", func(t *testing.T) {
		tests := []struct {
			input          string
			expectedParams []string
		}{
			{"fn() {}", []string{}},
			{"fn(x) {}", []string{"x"}},
			{"fn(x, y, z) {}", []string{"x", "y", "z"}},
		}

		for _, test := range tests {
			program := parseInput(t, test.input)

			statement := program.Statements[0].(*ast.ExpressionStatement)
			functionLiteral := statement.Expression.(*ast.FunctionLiteral)

			if len(functionLiteral.Parameters) != len(test.expectedParams) {
				t.Errorf("Expected %d function parameters, got %d", len(functionLiteral.Parameters), len(test.expectedParams))
			}

			for i, identifier := range test.expectedParams {
				assertLiteralExpression(t, functionLiteral.Parameters[i], identifier)
			}
		}
	})

	t.Run("Parse Call expressions", func(t *testing.T) {
		input := "add(1, 2 * 3, 4 + 5);"

		program := parseInput(t, input)
		assertStatementsPresent(t, program)

		expressionStatement, ok := program.Statements[0].(*ast.ExpressionStatement)
		assertNodeType(t, ok, expressionStatement, "*ast.ExpressionStatement")

		expression, ok := expressionStatement.Expression.(*ast.CallExpression)
		assertNodeType(t, ok, expression, "*ast.CallExpression")

		assertIdentifierLiteral(t, expression.Function, "add")
		assertArgumentLength(t, expression, 3)

		assertLiteralExpression(t, expression.Arguments[0], 1)
		assertInfixExpression(t, expression.Arguments[1], 2, "*", 3)
		assertInfixExpression(t, expression.Arguments[2], 4, "+", 5)
	})

	t.Run("Parse call arguments", func(t *testing.T) {
		tests := []struct {
			input              string
			expectedIdentifier string
			expectedArguments  []string
		}{
			{
				input:              "add();",
				expectedIdentifier: "add",
				expectedArguments:  []string{},
			},
			{
				input:              "add(1);",
				expectedIdentifier: "add",
				expectedArguments:  []string{"1"},
			},
			{
				input:              "add(1, 2 * 3, 4 + 5);",
				expectedIdentifier: "add",
				expectedArguments:  []string{"1", "(2 * 3)", "(4 + 5)"},
			},
		}

		for _, test := range tests {
			program := parseInput(t, test.input)

			expressionStatement, ok := program.Statements[0].(*ast.ExpressionStatement)
			assertNodeType(t, ok, expressionStatement, "*ast.ExpressionStatement")
			expression, ok := expressionStatement.Expression.(*ast.CallExpression)
			assertNodeType(t, ok, expression, "*ast.CallExpression")

			assertArgumentLength(t, expression, len(test.expectedArguments))

			for i, arg := range test.expectedArguments {
				if expression.Arguments[i].String() != arg {
					t.Errorf("Argument mismatch. Expected %q, got %q", arg, expression.Arguments[i])
				}
			}
		}
	})

	t.Run("Precedence", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{
				"-a * b",
				"((-a) * b)",
			},
			{
				"!-a",
				"(!(-a))",
			},
			{
				"a + b + c",
				"((a + b) + c)",
			},
			{
				"a + b - c",
				"((a + b) - c)",
			},
			{
				"a * b * c",
				"((a * b) * c)",
			},
			{
				"a * b / c",
				"((a * b) / c)",
			},
			{
				"a + b / c",
				"(a + (b / c))",
			},
			{
				"a + b * c + d / e - f",
				"(((a + (b * c)) + (d / e)) - f)",
			},
			{
				"3 + 4; -5 * 5",
				"(3 + 4)((-5) * 5)",
			},
			{
				"5 > 4 == 3 < 4",
				"((5 > 4) == (3 < 4))",
			},
			{
				"5 < 4 != 3 < 4",
				"((5 < 4) != (3 < 4))",
			},
			{
				"3 + 4 * 5 == 3 * 1 + 4 * 5",
				"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
			},
			{
				"true",
				"true",
			},
			{
				"false",
				"false",
			},
			{
				"3 > 5 == false",
				"((3 > 5) == false)",
			},
			{
				"3 < 5 == true",
				"((3 < 5) == true)",
			},
			{
				"1 + (2 + 3) +4",
				"((1 + (2 + 3)) + 4)",
			},
			{
				"(5 + 5) * 2",
				"((5 + 5) * 2)",
			},
			{
				"2 / (5 + 5)",
				"(2 / (5 + 5))",
			},
			{
				"-(5 + 5)",
				"(-(5 + 5))",
			},
			{
				"!(true == true)",
				"(!(true == true))",
			},
			{
				"a + add(b * c) + d",
				"((a + add((b * c))) + d)",
			},
			{
				"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
				"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
			},
			{
				"add(a + b + c * d / f + g)",
				"add((((a + b) + ((c * d) / f)) + g))",
			},
		}

		for _, test := range tests {
			program := parseInput(t, test.input)

			actual := program.String()

			if actual != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, actual)
			}
		}
	})
}

func assertStatementsPresent(t *testing.T, program *ast.Program) {
	t.Helper()

	if len(program.Statements) == 0 {
		t.Fatalf("No statements present")
	}
}

func assertLiteralExpression(t *testing.T, expression ast.Expression, expected interface{}) {
	t.Helper()

	switch expectedType := expected.(type) {
	case int:
		assertLiteralExpression(t, expression, int64(expectedType))
	case int64:
		assertIntegerLiteral(t, expression, expectedType)
	case string:
		assertIdentifierLiteral(t, expression, expectedType)
	case bool:
		assertBooleanLiteral(t, expression, expectedType)
	default:
		t.Fatalf("Type of expression not handled: %q", expectedType)
	}
}

func assertBlockExpression(t *testing.T, blockStatement *ast.BlockStatement, want interface{}) {
	t.Helper()

	if len(blockStatement.Statements) != 1 {
		t.Errorf("Block statements amount mismatch. Expected 1, got %d", len(blockStatement.Statements))
	}

	statement, ok := blockStatement.Statements[0].(*ast.ExpressionStatement)
	assertNodeType(t, ok, statement, "*ast.ExpressionStatement")

	if want != nil {
		assertLiteralExpression(t, statement.Expression, want)
	}
}

func assertInfixExpression(t *testing.T, expression ast.Expression, left interface{}, operator string, right interface{}) {
	t.Helper()
	infixExpression, ok := expression.(*ast.InfixExpression)

	assertNodeType(t, ok, infixExpression, "*ast.InfixExpression")
	assertLiteralExpression(t, infixExpression.Left, left)

	if infixExpression.Operator != operator {
		t.Errorf("Operator does not match. Expected %q, got %q", infixExpression.Operator, operator)
	}

	assertLiteralExpression(t, infixExpression.Right, right)
}

func assertIntegerLiteral(t *testing.T, expression ast.Expression, want int64) {
	t.Helper()
	integerLiteral, ok := expression.(*ast.IntegerLiteral)

	assertNodeType(t, ok, integerLiteral, "*ast.IntegerLiteral")
	assertIntegerLiteralValue(t, integerLiteral, want)
	assertTokenLiteral(t, integerLiteral, strconv.FormatInt(want, 10))
}

func assertIdentifierLiteral(t *testing.T, expression ast.Expression, want string) {
	t.Helper()
	identifier, ok := expression.(*ast.Identifier)

	assertNodeType(t, ok, identifier, "*ast.Identifier")
	assertIdentifierValue(t, identifier, want)
	assertTokenLiteral(t, identifier, want)
}

func assertBooleanLiteral(t *testing.T, expression ast.Expression, want bool) {
	t.Helper()
	boolean, ok := expression.(*ast.BooleanLiteral)

	assertNodeType(t, ok, boolean, "*ast.Boolean")
	assertBooleanValue(t, boolean, want)
}

func assertIdentifierValue(t *testing.T, identifier *ast.Identifier, want interface{}) {
	t.Helper()

	if identifier.Value != want {
		t.Errorf("Identifier value mismatch. Expected %q, got %q", want, identifier.Value)
	}
}

func assertIntegerLiteralValue(t *testing.T, integerLiteral *ast.IntegerLiteral, want int64) {
	t.Helper()

	if integerLiteral.Value != want {
		t.Errorf("IntegerLiteral value mismatch. Expected %q, got %q", want, integerLiteral.Value)
	}
}

func assertBooleanValue(t *testing.T, booleanLiteral *ast.BooleanLiteral, want bool) {
	t.Helper()

	if booleanLiteral.Value != want {
		t.Errorf("Boolean valie mismatch. Expected %t, got %t", booleanLiteral.Value, want)
	}
}

func assertArgumentLength(t *testing.T, callExpression *ast.CallExpression, want int) {
	t.Helper()

	if len(callExpression.Arguments) != want {
		t.Errorf("Expected %d arguments, got %d", want, len(callExpression.Arguments))
	}
}

func assertOperator(t *testing.T, operator, want string) {
	t.Helper()

	if operator != want {
		t.Errorf("Operator mismatch. Expected %q, got %q", want, operator)
	}
}

func assertIdentifierTokenLiteral(t *testing.T, identifier *ast.Identifier, want string) {
	t.Helper()

	if identifier.TokenLiteral() != want {
		t.Errorf("Identifier TokenValue mismatch. Expected %q, got %q", want, identifier.TokenLiteral())
	}
}

func assertNodeType(t *testing.T, ok bool, node ast.Node, want string) {
	t.Helper()

	if ok {
		return
	}

	t.Errorf("Type of Node not correct. Expected %q, got %T", want, node)
}

func assertTokenLiteral(t *testing.T, node ast.Node, want interface{}) {
	t.Helper()

	if node.TokenLiteral() != want {
		t.Errorf("TokenLiteral mismatch. Expected %q, got %q", want, node.TokenLiteral())
	}
}

func parseInput(t *testing.T, input string) *ast.Program {
	t.Helper()

	lexer := lexer.NewLexer(input)
	parser := NewParser(lexer)

	program := parser.ParseProgram()

	if len(parser.Errors()) != 0 {
		for _, error := range parser.Errors() {
			t.Error("Parser error:", error)
		}

		t.FailNow()
	}

	if program == nil {
		t.Fatalf("No program present")
	}

	return program
}
