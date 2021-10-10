package token

type TokenType string

type Token struct {
	StartChar byte // 被视为非法token字面量的一部分，不会被视为合法token的一部分，主要是为了便于追踪错误
	Type TokenType
	Literal string
}

// token types
const (
	ILLEGAL = "ILLEGAL"
	EOF 	= "EOF"

	//Identifiers and literals
	IDENT	= "IDENT"
	INT 	= "INT"

	//Operators
	ASSIGN 	= "="
	PLUS	= "+"
	MINUS	= "-"
	BANG	= "!"
	ASTERISK = "*"
	SLASH	= "/"

	LT		= "<"
	GT		= ">"
	EQ		= "=="
	NOT_EQ	= "!="

	//delimiters
	COMMA	= ","
	SEMICOLON	= ";"
	LPAREN	= "("
	RPAREN	= ")"
	LBRACE	= "{"
	RBRACE	= "}"
	LBRACKET = "["
	RBRACKET = "]"

	//keywords
	FUNCTION = "FUNCTION"
	LET		= "LET"
	TRUE = "TRUE"
	FALSE = "FALSE"
	IF = "IF"
	ELSE = "ELSE"
	RETURN = "RETURN"
	STRING = "STRING"
)

var keywords = map[string]TokenType{
	"fn": FUNCTION,
	"let": LET,
	"true": TRUE,
	"false": FALSE,
	"if": IF,
	"else": ELSE,
	"return": RETURN,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

func (tok Token) isLegal() bool {
	if tok.Type != ILLEGAL {
		return true
	}else {
		return false
	}
}
