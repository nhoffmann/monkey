package parser

import (
	"fmt"
	"strconv"

	"github.com/nhoffmann/monkey/ast"
	"github.com/nhoffmann/monkey/lexer"
	"github.com/nhoffmann/monkey/token"
)

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Parser provides functionality to turn a given input in to an AST
type Parser struct {
	lexer  *lexer.Lexer
	errors []error

	currentToken token.Token
	peekToken    token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

// NewParser creates a new parser instance
func NewParser(l *lexer.Lexer) *Parser {
	parser := &Parser{
		lexer:  l,
		errors: []error{},
	}

	parser.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	parser.registerPrefix(token.IDENT, parser.parseIdentifier)
	parser.registerPrefix(token.INT, parser.parseIntegerLiteral)
	parser.registerPrefix(token.BANG, parser.parsePrefixExpression)
	parser.registerPrefix(token.MINUS, parser.parsePrefixExpression)

	parser.infixParseFns = make(map[token.TokenType]infixParseFn)
	parser.registerInfix(token.PLUS, parser.parseInfixExpression)
	parser.registerInfix(token.MINUS, parser.parseInfixExpression)
	parser.registerInfix(token.ASTERISK, parser.parseInfixExpression)
	parser.registerInfix(token.SLASH, parser.parseInfixExpression)
	parser.registerInfix(token.LT, parser.parseInfixExpression)
	parser.registerInfix(token.GT, parser.parseInfixExpression)
	parser.registerInfix(token.EQ, parser.parseInfixExpression)
	parser.registerInfix(token.NOT_EQ, parser.parseInfixExpression)

	parser.nextToken()
	parser.nextToken()

	return parser
}

// ParseProgram builds an AST from a given input
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.currentTokenIs(token.EOF) {
		statement := p.parseStatement()
		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}
		p.nextToken()
	}

	return program
}

// Errors returns all parsing errors
func (p *Parser) Errors() []error {
	return p.errors
}

func (p *Parser) registerPrefix(token token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[token] = fn
}

func (p *Parser) registerInfix(token token.TokenType, fn infixParseFn) {
	p.infixParseFns[token] = fn
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.currentToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	statement := &ast.LetStatement{Token: p.currentToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	statement.Name = &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	if !p.currentTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return statement
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	statement := &ast.ReturnStatement{Token: p.currentToken}

	p.nextToken()

	if !p.currentTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return statement
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	statement := &ast.ExpressionStatement{Token: p.currentToken}

	statement.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return statement
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.currentToken.Type]

	if prefix == nil {
		p.registerParseError(&NoPrefixParseFunctionError{p.currentToken.Type})
		return nil
	}

	leftExpression := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]

		if infix == nil {
			return leftExpression
		}

		p.nextToken()

		leftExpression = infix(leftExpression)
	}

	return leftExpression
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	literal := &ast.IntegerLiteral{Token: p.currentToken}

	value, err := strconv.ParseInt(p.currentToken.Literal, 0, 64)
	if err != nil {
		p.registerParseError(&UnparsableIntegerError{p.currentToken.Literal})
		return nil
	}

	literal.Value = value

	return literal
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.currentToken,
		Operator: p.currentToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.currentToken,
		Operator: p.currentToken.Literal,
		Left:     left,
	}

	precedence := p.currentPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) peekPrecedence() int {
	if precedence, ok := precedences[p.peekToken.Type]; ok {
		return precedence
	}

	return LOWEST
}

func (p *Parser) currentPrecedence() int {
	if precedence, ok := precedences[p.currentToken.Type]; ok {
		return precedence
	}

	return LOWEST
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) currentTokenIs(tokenType token.TokenType) bool {
	return p.currentToken.Type == tokenType
}

func (p *Parser) peekTokenIs(tokenType token.TokenType) bool {
	return p.peekToken.Type == tokenType
}

func (p *Parser) expectPeek(tokenType token.TokenType) bool {
	if !p.peekTokenIs(tokenType) {
		p.registerParseError(&PeekError{tokenType, p.peekToken.Type})
		return false
	}

	p.nextToken()
	return true
}

func (p *Parser) registerParseError(error error) {
	p.errors = append(p.errors, error)
}

type PeekError struct {
	expectedTokenType token.TokenType
	actualTokenType   token.TokenType
}

func (pe *PeekError) Error() string {
	return fmt.Sprintf("Expected next token to be %q, but got %q", pe.expectedTokenType, pe.actualTokenType)
}

type UnparsableIntegerError struct {
	literal string
}

func (uie *UnparsableIntegerError) Error() string {
	return fmt.Sprintf("Counld not parse input to integer: %q", uie.literal)
}

type NoPrefixParseFunctionError struct {
	tokenType token.TokenType
}

func (nppfe *NoPrefixParseFunctionError) Error() string {
	return fmt.Sprintf("No prefixParseFunction for given token: %q", nppfe.tokenType)
}
