package parser

import (
	"fmt"
	"junk/ast"
	"junk/lexer"
	"junk/token"
	"strconv"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

const (
	_ int = iota // ignore first value by assigning to blank identifier
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn // map of prefix parse functions
	infixParseFns  map[token.TokenType]infixParseFn  // map of infix parse functions
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn) // initialize map
	p.registerPrefix(token.IDENT, p.parseIdentifier)           // register identifier parse function
	p.registerPrefix(token.INT, p.parseIntegerLiteral)         // register integer literal parse function

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal} // initialize identifier
}

func (p *Parser) Errors() []string { // return errors
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken() // read next token
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{} // initialize empty slice

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt) // append to slice
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type { // check current token type
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement() // parse expression statement
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement { // parse let statement
	stmt := &ast.LetStatement{Token: p.curToken} // initialize let statement

	if !p.expectPeek(token.IDENT) { // check next token type
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal} // initialize identifier

	if !p.expectPeek(token.ASSIGN) { // check next token type
		return nil
	}

	// TODO : We're skipping expressions until we encounter a semicolon

	for !p.curTokenIs(token.SEMICOLON) { // check current token type
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement { // parse return statement
	stmt := &ast.ReturnStatement{Token: p.curToken} // initialize return statement

	p.nextToken()

	// TODO : We're skipping expressions until we encounter a semicolon

	for !p.curTokenIs(token.SEMICOLON) { // check current token type
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement { // parse expression statement
	stmt := &ast.ExpressionStatement{Token: p.curToken} // initialize expression statement

	stmt.Expression = p.parseExpression(LOWEST) // parse expression

	if p.peekTokenIs(token.SEMICOLON) { // check next token type
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type] // get prefix parse function
	if prefix == nil {                          // check if prefix parse function exists
		return nil
	}
	leftExp := prefix() // parse prefix expression

	return leftExp
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) curTokenIs(t token.TokenType) bool { // check current token type
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool { // check next token type
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool { // check next token type
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t) // add error
		return false
	}
}

func (p *Parser) peekError(t token.TokenType) { // add error
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
