package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType string

const (
	// Literals
	TOKEN_INT_LIT    TokenType = "INT_LIT"
	TOKEN_FLOAT_LIT  TokenType = "FLOAT_LIT"
	TOKEN_STRING_LIT TokenType = "STRING_LIT"
	TOKEN_BOOL_LIT   TokenType = "BOOL_LIT"

	// Types
	TOKEN_TYPE_INT    TokenType = "TYPE_INT"
	TOKEN_TYPE_FLOAT  TokenType = "TYPE_FLOAT"
	TOKEN_TYPE_STRING TokenType = "TYPE_STRING"
	TOKEN_TYPE_BOOL   TokenType = "TYPE_BOOL"

	// Keywords
	TOKEN_IF   TokenType = "IF"
	TOKEN_STOP TokenType = "STOP"
	TOKEN_LIB  TokenType = "LIB"
	TOKEN_IMPORT TokenType = "IMPORT"
	TOKEN_PRINT TokenType = "PRINT"

	// Identifiers
	TOKEN_IDENT TokenType = "IDENT"

	// Operators
	TOKEN_ASSIGN TokenType = "ASSIGN"
	TOKEN_PLUS   TokenType = "PLUS"
	TOKEN_MINUS  TokenType = "MINUS"
	TOKEN_STAR   TokenType = "STAR"
	TOKEN_SLASH  TokenType = "SLASH"
	TOKEN_EQ     TokenType = "EQ"
	TOKEN_NEQ    TokenType = "NEQ"
	TOKEN_LT     TokenType = "LT"
	TOKEN_LTE    TokenType = "LTE"
	TOKEN_GT     TokenType = "GT"
	TOKEN_GTE    TokenType = "GTE"

	// Punctuation
	TOKEN_LPAREN    TokenType = "LPAREN"
	TOKEN_RPAREN    TokenType = "RPAREN"
	TOKEN_SEMICOLON TokenType = "SEMICOLON"
	TOKEN_DOT       TokenType = "DOT"
	TOKEN_COMMA     TokenType = "COMMA"

	// Special
	TOKEN_EOF     TokenType = "EOF"
	TOKEN_NEWLINE TokenType = "NEWLINE"
)

type Token struct {
	Type    TokenType
	Value   string
	Line    int
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%s, %q, line=%d)", t.Type, t.Value, t.Line)
}

var keywords = map[string]TokenType{
	"int":    TOKEN_TYPE_INT,
	"float":  TOKEN_TYPE_FLOAT,
	"string": TOKEN_TYPE_STRING,
	"bool":   TOKEN_TYPE_BOOL,
	"if":     TOKEN_IF,
	"stop":   TOKEN_STOP,
	"lib":    TOKEN_LIB,
	"import": TOKEN_IMPORT,
	"print":  TOKEN_PRINT,
	"true":   TOKEN_BOOL_LIT,
	"false":  TOKEN_BOOL_LIT,
}

type Lexer struct {
	src    []rune
	pos    int
	line   int
	tokens []Token
}

func New(src string) *Lexer {
	return &Lexer{src: []rune(src), pos: 0, line: 1}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *Lexer) peekAt(offset int) rune {
	idx := l.pos + offset
	if idx >= len(l.src) {
		return 0
	}
	return l.src[idx]
}

func (l *Lexer) advance() rune {
	ch := l.src[l.pos]
	l.pos++
	return ch
}

func (l *Lexer) skipComment() {
	// Comments start with $/
	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.pos++
	}
}

func (l *Lexer) readString() Token {
	l.pos++ // skip opening "
	var sb strings.Builder
	for l.pos < len(l.src) && l.src[l.pos] != '"' {
		sb.WriteRune(l.advance())
	}
	l.pos++ // skip closing "
	return Token{Type: TOKEN_STRING_LIT, Value: sb.String(), Line: l.line}
}

func (l *Lexer) readNumber() Token {
	var sb strings.Builder
	isFloat := false
	for l.pos < len(l.src) && (unicode.IsDigit(l.src[l.pos]) || l.src[l.pos] == '.') {
		if l.src[l.pos] == '.' {
			isFloat = true
		}
		sb.WriteRune(l.advance())
	}
	if isFloat {
		return Token{Type: TOKEN_FLOAT_LIT, Value: sb.String(), Line: l.line}
	}
	return Token{Type: TOKEN_INT_LIT, Value: sb.String(), Line: l.line}
}

func (l *Lexer) readIdent() Token {
	var sb strings.Builder
	for l.pos < len(l.src) && (unicode.IsLetter(l.src[l.pos]) || unicode.IsDigit(l.src[l.pos]) || l.src[l.pos] == '_') {
		sb.WriteRune(l.advance())
	}
	val := sb.String()
	if tt, ok := keywords[val]; ok {
		return Token{Type: tt, Value: val, Line: l.line}
	}
	return Token{Type: TOKEN_IDENT, Value: val, Line: l.line}
}

func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token

	for l.pos < len(l.src) {
		ch := l.peek()

		// Skip spaces and tabs
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advance()
			continue
		}

		// Newlines
		if ch == '\n' {
			tokens = append(tokens, Token{Type: TOKEN_NEWLINE, Value: "\\n", Line: l.line})
			l.line++
			l.advance()
			continue
		}

		// Comments: $/
		if ch == '$' && l.peekAt(1) == '/' {
			l.skipComment()
			continue
		}

		// String literals
		if ch == '"' {
			tokens = append(tokens, l.readString())
			continue
		}

		// Numbers
		if unicode.IsDigit(ch) {
			tokens = append(tokens, l.readNumber())
			continue
		}

		// Identifiers and keywords
		if unicode.IsLetter(ch) || ch == '_' {
			tokens = append(tokens, l.readIdent())
			continue
		}

		// Two-char operators
		if ch == '<' && l.peekAt(1) == '=' {
			tokens = append(tokens, Token{Type: TOKEN_LTE, Value: "<=", Line: l.line})
			l.pos += 2
			continue
		}
		if ch == '>' && l.peekAt(1) == '=' {
			tokens = append(tokens, Token{Type: TOKEN_GTE, Value: ">=", Line: l.line})
			l.pos += 2
			continue
		}
		if ch == '!' && l.peekAt(1) == '=' {
			tokens = append(tokens, Token{Type: TOKEN_NEQ, Value: "!=", Line: l.line})
			l.pos += 2
			continue
		}
		if ch == '=' && l.peekAt(1) == '=' {
			tokens = append(tokens, Token{Type: TOKEN_EQ, Value: "==", Line: l.line})
			l.pos += 2
			continue
		}

		// Single-char tokens
		switch ch {
		case '=':
			tokens = append(tokens, Token{Type: TOKEN_ASSIGN, Value: "=", Line: l.line})
		case '+':
			tokens = append(tokens, Token{Type: TOKEN_PLUS, Value: "+", Line: l.line})
		case '-':
			tokens = append(tokens, Token{Type: TOKEN_MINUS, Value: "-", Line: l.line})
		case '*':
			tokens = append(tokens, Token{Type: TOKEN_STAR, Value: "*", Line: l.line})
		case '/':
			tokens = append(tokens, Token{Type: TOKEN_SLASH, Value: "/", Line: l.line})
		case '<':
			tokens = append(tokens, Token{Type: TOKEN_LT, Value: "<", Line: l.line})
		case '>':
			tokens = append(tokens, Token{Type: TOKEN_GT, Value: ">", Line: l.line})
		case '(':
			tokens = append(tokens, Token{Type: TOKEN_LPAREN, Value: "(", Line: l.line})
		case ')':
			tokens = append(tokens, Token{Type: TOKEN_RPAREN, Value: ")", Line: l.line})
		case ';':
			tokens = append(tokens, Token{Type: TOKEN_SEMICOLON, Value: ";", Line: l.line})
		case '.':
			tokens = append(tokens, Token{Type: TOKEN_DOT, Value: ".", Line: l.line})
		case ',':
			tokens = append(tokens, Token{Type: TOKEN_COMMA, Value: ",", Line: l.line})
		default:
			return nil, fmt.Errorf("line %d: unexpected character %q", l.line, ch)
		}
		l.advance()
	}

	tokens = append(tokens, Token{Type: TOKEN_EOF, Value: "", Line: l.line})
	return tokens, nil
}
