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
	INDEX       // array[index]
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn // map of prefix parse functions
	infixParseFns  map[token.TokenType]infixParseFn  // map of infix parse functions
}

var precedences = map[token.TokenType]int{ // map of precedences
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn) // initialize map
	p.registerPrefix(token.IDENT, p.parseIdentifier)           // register identifier parse function
	p.registerPrefix(token.INT, p.parseIntegerLiteral)         // register integer literal parse function
	p.registerPrefix(token.BANG, p.parsePrefixExpression)      // register prefix expression parse function
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)     // register prefix expression parse function
	p.registerPrefix(token.TRUE, p.parseBoolean)               // register boolean parse function
	p.registerPrefix(token.FALSE, p.parseBoolean)              // register boolean parse function
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)   // register grouped expression parse function
	p.registerPrefix(token.IF, p.parseIfExpression)            // register if expression parse function
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)   // register function literal parse function
	p.registerPrefix(token.STRING, p.parseStringLiteral)       // register string literal parse function
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)      // register array literal parse function
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)         // register hash literal parse function
	p.registerPrefix(token.MACRO, p.parseMacroLiteral)         // register macro literal parse function

	p.infixParseFns = make(map[token.TokenType]infixParseFn) // initialize map
	p.registerInfix(token.PLUS, p.parseInfixExpression)      // register infix expression parse function
	p.registerInfix(token.MINUS, p.parseInfixExpression)     // register infix expression parse function
	p.registerInfix(token.SLASH, p.parseInfixExpression)     // register infix expression parse function
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)  // register infix expression parse function
	p.registerInfix(token.EQ, p.parseInfixExpression)        // register infix expression parse function
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)    // register infix expression parse function
	p.registerInfix(token.LT, p.parseInfixExpression)        // register infix expression parse function
	p.registerInfix(token.GT, p.parseInfixExpression)        // register infix expression parse function
	p.registerInfix(token.LPAREN, p.parseCallExpression)     // register call expression parse function
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)  // register index expression parse function

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal} // initialize identifier
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)} // initialize boolean
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken} // initialize if expression

	if !p.expectPeek(token.LPAREN) { // check next token type
		return nil
	}

	p.nextToken()

	expression.Condition = p.parseExpression(LOWEST) // parse expression

	if !p.expectPeek(token.RPAREN) { // check next token type
		return nil
	}

	if !p.expectPeek(token.LBRACE) { // check next token type
		return nil
	}

	expression.Consequence = p.parseBlockStatement() // parse block statement

	if p.peekTokenIs(token.ELSE) { // check next token type
		p.nextToken()

		if !p.expectPeek(token.LBRACE) { // check next token type
			return nil
		}

		expression.Alternative = p.parseBlockStatement() // parse block statement
	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken} // initialize block statement
	block.Statements = []ast.Statement{}            // initialize empty slice

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) { // check current token type
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt) // append to slice
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST) // parse expression

	if !p.expectPeek(token.RPAREN) { // check next token type
		return nil
	}

	return exp
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken} // initialize function literal

	if !p.expectPeek(token.LPAREN) { // check next token type
		return nil
	}

	lit.Parameters = p.parseFunctionParameters() // parse function parameters

	if !p.expectPeek(token.LBRACE) { // check next token type
		return nil
	}

	lit.Body = p.parseBlockStatement() // parse block statement

	return lit
}

func (p *Parser) parseMacroLiteral() ast.Expression {
	lit := &ast.MacroLiteral{Token: p.curToken} // initialize macro literal

	if !p.expectPeek(token.LPAREN) { // check next token type
		return nil
	}

	lit.Parameters = p.parseFunctionParameters() // parse function parameters

	if !p.expectPeek(token.LBRACE) { // check next token type
		return nil
	}

	lit.Body = p.parseBlockStatement() // parse block statement

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function} // initialize call expression
	exp.Arguments = p.parseCallArguments()                            // parse expression list
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) { // check next token type
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST)) // parse expression

	for p.peekTokenIs((token.COMMA)) { // check next token type
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal} // initialize string literal
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}          // initialize array literal
	array.Elements = p.parseExpressionList(token.RBRACKET) // parse expression list
	return array
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken} // initialize hash literal
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.RBRACE) { // check next token type
		p.nextToken()
		key := p.parseExpression(LOWEST) // parse expression

		if !p.expectPeek(token.COLON) { // check next token type
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST) // parse expression

		hash.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) { // check next token type
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) { // check next token type
		return nil
	}

	return hash
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST)) // parse expression

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left} // initialize index expression

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST) // parse expression

	if !p.expectPeek(token.RBRACKET) { // check next token type
		return nil
	}

	return exp
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
	case token.WHILE:
		return p.parseWhileStatement()
	default:
		return p.parseExpressionStatement() // parse expression statement
	}
}

func (p *Parser) parseWhileStatement() ast.Statement {
	statement := &ast.WhileStatement{
		Token: p.curToken,
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	statement.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	statement.Body = p.parseBlockStatement()

	return statement
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

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST) // parse expression

	for p.peekTokenIs(token.SEMICOLON) { // check current token type
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement { // parse return statement
	stmt := &ast.ReturnStatement{Token: p.curToken} // initialize return statement

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	for p.peekTokenIs(token.SEMICOLON) { // check current token type
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
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix() // parse prefix expression

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() { // check next token type
		infix := p.infixParseFns[p.peekToken.Type] // get infix parse function
		if infix == nil {                          // check if infix parse function exists
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp) // parse infix expression
	}

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

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	} // initialize prefix expression

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{ // initialize infix expression
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	} // initialize infix expression

	precedence := p.curPrecedence() // get precedence of current token
	p.nextToken()

	expression.Right = p.parseExpression(precedence) // parse expression

	return expression
}

func (p *Parser) peekPrecedence() int { // get precedence of next token
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int { // get precedence of current token
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}
