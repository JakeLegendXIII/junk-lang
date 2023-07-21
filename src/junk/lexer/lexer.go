package lexer

import (
	"junk/token"
)

type Lexer struct {
	input        string
	postion      int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // channel of chars being read
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.eatWhitespace() // helper function

	switch l.ch {
	case '=':
		tok = newToken(token.ASSIGN, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) { // isLetter is a helper function
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal) // LookupIdent is a helper function
			return tok
		} else if isDigit(l.ch) { // isDigit is a helper function
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII code for "NUL" (null character)
	} else {
		l.ch = l.input[l.readPosition]
	}

	l.postion = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) readNumber() string { // helper function
	position := l.postion
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.postion]
}

func (l *Lexer) readIdentifier() string { // helper function
	position := l.postion
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.postion]
}

func isLetter(ch byte) bool { // helper function
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '!' || ch == '?'
}

func isDigit(ch byte) bool { // helper function
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) eatWhitespace() { // helper function
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}
