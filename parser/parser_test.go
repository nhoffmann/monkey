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
			// expectedValue      interface{}
		}{
			// {"let x = 5;", "x"},
			// {"let y = true;", "y"},
			// {"let foobar = y;", "foobar"},
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
