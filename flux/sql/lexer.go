package sql

import (
	"strings"
	"unicode"
)

type TokenType string

const (
	TokenIllegal TokenType = "ILLEGAL"
	TokenEOF     TokenType = "EOF"
	TokenIdent   TokenType = "IDENT"
	TokenInt     TokenType = "INT"
	TokenFloat   TokenType = "FLOAT"
	TokenString  TokenType = "STRING"

	// Keywords
	TokenSelect TokenType = "SELECT"
	TokenFrom   TokenType = "FROM"
	TokenWhere  TokenType = "WHERE"
	TokenAnd    TokenType = "AND"
	TokenJoin   TokenType = "JOIN"
	TokenOn     TokenType = "ON"
	TokenOrder  TokenType = "ORDER"
	TokenBy     TokenType = "BY"
	TokenLimit  TokenType = "LIMIT"
	TokenAsc    TokenType = "ASC"
	TokenDesc   TokenType = "DESC"
	TokenCreate TokenType = "CREATE"
	TokenTable  TokenType = "TABLE"
	TokenInsert TokenType = "INSERT"
	TokenInto   TokenType = "INTO"
	TokenValues TokenType = "VALUES"

	// Vector functions
	TokenVectorSim  TokenType = "VECTOR_SIMILARITY"
	TokenVectorDist TokenType = "VECTOR_DISTANCE"

	// Operators and Delimiters
	TokenAssign TokenType = "="
	TokenGT     TokenType = ">"
	TokenLT     TokenType = "<"
	TokenComma  TokenType = ","
	TokenLParen TokenType = "("
	TokenRParen TokenType = ")"
	TokenDot    TokenType = "."
	TokenSemi   TokenType = ";"
)

type Token struct {
	Type    TokenType
	Literal string
}

func Lex(input string) []Token {
	var tokens []Token
	runes := []rune(input)
	pos := 0

	for pos < len(runes) {
		ch := runes[pos]

		if unicode.IsSpace(ch) {
			pos++
			continue
		}

		if ch == '-' && pos+1 < len(runes) && runes[pos+1] == '-' {
			// comment
			for pos < len(runes) && runes[pos] != '\n' {
				pos++
			}
			continue
		}

		switch ch {
		case '=':
			tokens = append(tokens, Token{Type: TokenAssign, Literal: "="})
			pos++
		case '>':
			tokens = append(tokens, Token{Type: TokenGT, Literal: ">"})
			pos++
		case '<':
			tokens = append(tokens, Token{Type: TokenLT, Literal: "<"})
			pos++
		case ',':
			tokens = append(tokens, Token{Type: TokenComma, Literal: ","})
			pos++
		case '(':
			tokens = append(tokens, Token{Type: TokenLParen, Literal: "("})
			pos++
		case ')':
			tokens = append(tokens, Token{Type: TokenRParen, Literal: ")"})
			pos++
		case '.':
			tokens = append(tokens, Token{Type: TokenDot, Literal: "."})
			pos++
		case ';':
			tokens = append(tokens, Token{Type: TokenSemi, Literal: ";"})
			pos++
		case '\'':
			// string
			pos++
			start := pos
			for pos < len(runes) && runes[pos] != '\'' {
				pos++
			}
			tokens = append(tokens, Token{Type: TokenString, Literal: string(runes[start:pos])})
			if pos < len(runes) {
				pos++ // skip closing quote
			}
		default:
			if unicode.IsLetter(ch) || ch == '_' {
				start := pos
				for pos < len(runes) && (unicode.IsLetter(runes[pos]) || unicode.IsDigit(runes[pos]) || runes[pos] == '_') {
					pos++
				}
				ident := string(runes[start:pos])
				tokens = append(tokens, Token{Type: lookupIdent(ident), Literal: ident})
			} else if unicode.IsDigit(ch) {
				start := pos
				isFloat := false
				for pos < len(runes) && (unicode.IsDigit(runes[pos]) || runes[pos] == '.') {
					if runes[pos] == '.' {
						isFloat = true
					}
					pos++
				}
				if isFloat {
					tokens = append(tokens, Token{Type: TokenFloat, Literal: string(runes[start:pos])})
				} else {
					tokens = append(tokens, Token{Type: TokenInt, Literal: string(runes[start:pos])})
				}
			} else {
				tokens = append(tokens, Token{Type: TokenIllegal, Literal: string(ch)})
				pos++
			}
		}
	}

	tokens = append(tokens, Token{Type: TokenEOF, Literal: ""})
	return tokens
}

var keywords = map[string]TokenType{
	"SELECT":            TokenSelect,
	"FROM":              TokenFrom,
	"WHERE":             TokenWhere,
	"AND":               TokenAnd,
	"JOIN":              TokenJoin,
	"ON":                TokenOn,
	"ORDER":             TokenOrder,
	"BY":                TokenBy,
	"LIMIT":             TokenLimit,
	"ASC":               TokenAsc,
	"DESC":              TokenDesc,
	"CREATE":            TokenCreate,
	"TABLE":             TokenTable,
	"INSERT":            TokenInsert,
	"INTO":              TokenInto,
	"VALUES":            TokenValues,
	"VECTOR_SIMILARITY": TokenVectorSim,
	"VECTOR_DISTANCE":   TokenVectorDist,
}

func lookupIdent(ident string) TokenType {
	if tok, ok := keywords[strings.ToUpper(ident)]; ok {
		return tok
	}
	return TokenIdent
}
