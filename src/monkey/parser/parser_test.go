package parser

import (
	"fmt"
	"strconv"
	"testing"

	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
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

	t.Run("Parse prefix expression", func(t *testing.T) {
		tests := []struct {
			input        string
			operator     string
			integerValue int64
		}{
			{"!5", "!", 5},
			{"-15", "-", 15},
		}

		for _, test := range tests {
			program := parseInput(t, test.input)

			assertStatementsPresent(t, program)

			expressionStatement, ok := program.Statements[0].(*ast.ExpressionStatement)

			assertNodeType(t, ok, expressionStatement, "*ast.ExpressionStatement")

			prefix, ok := expressionStatement.Expression.(*ast.PrefixExpression)

			assertNodeType(t, ok, prefix, "*ast.PrefixExpression")
			assertOperator(t, prefix.Operator, test.operator)

			assertIntegerLiteral(t, prefix.Right, test.integerValue)
		}
	})

	t.Run("Parse infix expressions", func(t *testing.T) {
		tests := []struct {
			input      string
			leftValue  int64
			operator   string
			rightValue int64
		}{
			{"5 + 5;", 5, "+", 5},
			{"5 - 5;", 5, "-", 5},
			{"5 * 5;", 5, "*", 5},
			{"5 / 5;", 5, "/", 5},
			{"5 > 5;", 5, ">", 5},
			{"5 < 5;", 5, "<", 5},
			{"5 == 5;", 5, "==", 5},
			{"5 != 5;", 5, "!=", 5},
		}

		for _, test := range tests {
			program := parseInput(t, test.input)

			assertStatementsPresent(t, program)

			expressionStatement, ok := program.Statements[0].(*ast.ExpressionStatement)

			assertNodeType(t, ok, expressionStatement, "*ast.ExpressionStatement")

			infix, ok := expressionStatement.Expression.(*ast.InfixExpression)

			assertNodeType(t, ok, infix, "*ast.InfixExpression")
			assertIntegerLiteral(t, infix.Left, test.leftValue)
			assertOperator(t, infix.Operator, test.operator)
			assertIntegerLiteral(t, infix.Right, test.rightValue)
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

func assertIntegerLiteral(t *testing.T, integerLiteral ast.Expression, want int64) {
	literal, ok := integerLiteral.(*ast.IntegerLiteral)

	assertNodeType(t, ok, literal, "*ast.IntegerLiteral")
	assertIntegerLiteralValue(t, literal, want)
	assertTokenLiteral(t, literal, strconv.FormatInt(want, 10))
}

func assertIdentifierValue(t *testing.T, identifier *ast.Identifier, want string) {
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

func assertTokenLiteral(t *testing.T, node ast.Node, want string) {
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
