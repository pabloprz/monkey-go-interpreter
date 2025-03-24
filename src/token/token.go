package token

type TokenType byte

type Token struct {
	Literal string
	Type    TokenType
}

const (
	ILLEGAL TokenType = iota
	EOF

	// Identifiers + literals
	IDENTIFIER
	INT

	// Operators
	ASSIGN
	PLUS

	// Delimeters
	COMMA
	SEMICOLON

	LPAREN
	RPAREN
	LBRACE
	RBRACE

	// Keywords
	FUNCTION
	LET
)

var keywords = map[string]TokenType{
	"fn":  FUNCTION,
	"let": LET,
}

func LookupIdentifier(identifier string) TokenType {
	if tok, ok := keywords[identifier]; ok {
		return tok
	}
	return IDENTIFIER
}
