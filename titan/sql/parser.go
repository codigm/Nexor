package sql

import (
	"fmt"
	"strconv"
)

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) peek() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{Type: TokenEOF, Literal: ""}
}

func (p *Parser) next() Token {
	tok := p.peek()
	p.pos++
	return tok
}

func (p *Parser) expect(tokType TokenType) (Token, error) {
	tok := p.next()
	if tok.Type != tokType {
		return tok, fmt.Errorf("expected %v, got %v (%s)", tokType, tok.Type, tok.Literal)
	}
	return tok, nil
}

func (p *Parser) ParseStatement() (Statement, error) {
	tok := p.peek()
	switch tok.Type {
	case TokenSelect:
		return p.parseSelect()
	case TokenCreate:
		return p.parseCreate()
	case TokenInsert:
		return p.parseInsert()
	default:
		return nil, fmt.Errorf("unexpected token %s", tok.Literal)
	}
}

func (p *Parser) parseSelect() (*SelectStatement, error) {
	p.expect(TokenSelect)
	stmt := &SelectStatement{}

	// columns
	for {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		stmt.Columns = append(stmt.Columns, expr)

		if p.peek().Type == TokenComma {
			p.next()
		} else {
			break
		}
	}

	// FROM
	if _, err := p.expect(TokenFrom); err != nil {
		return nil, err
	}
	tableTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}
	stmt.From = &TableIdent{Name: tableTok.Literal}
	
	// Alias
	if p.peek().Type == TokenIdent {
		aliasTok := p.next()
		stmt.From.Alias = aliasTok.Literal
	}

	// JOIN
	if p.peek().Type == TokenJoin {
		p.next()
		joinTable, err := p.expect(TokenIdent)
		if err != nil {
			return nil, err
		}
		joinAlias := ""
		if p.peek().Type == TokenIdent && p.peek().Type != TokenOn {
			joinAliasTok := p.next()
			joinAlias = joinAliasTok.Literal
		}
		_, err = p.expect(TokenOn)
		if err != nil {
			return nil, err
		}
		onExpr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		stmt.Join = &JoinClause{
			Table: &TableIdent{Name: joinTable.Literal, Alias: joinAlias},
			On:    onExpr,
		}
	}

	// WHERE
	if p.peek().Type == TokenWhere {
		p.next()
		whereExpr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		stmt.Where = whereExpr
	}

	// ORDER BY
	if p.peek().Type == TokenOrder {
		p.next()
		if _, err := p.expect(TokenBy); err != nil {
			return nil, err
		}
		orderExpr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		asc := true
		if p.peek().Type == TokenDesc {
			p.next()
			asc = false
		} else if p.peek().Type == TokenAsc {
			p.next()
		}
		stmt.OrderBy = &OrderByClause{Expr: orderExpr, Asc: asc}
	}

	// LIMIT
	if p.peek().Type == TokenLimit {
		p.next()
		limitTok, err := p.expect(TokenInt)
		if err != nil {
			return nil, err
		}
		val, _ := strconv.ParseInt(limitTok.Literal, 10, 64)
		stmt.Limit = &IntegerLiteral{Value: val}
	}

	if p.peek().Type == TokenSemi {
		p.next()
	}

	return stmt, nil
}

func (p *Parser) parseCreate() (*CreateTableStatement, error) {
	p.expect(TokenCreate)
	p.expect(TokenTable)
	nameTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}

	p.expect(TokenLParen)
	stmt := &CreateTableStatement{TableName: nameTok.Literal}
	
	for {
		colName, err := p.expect(TokenIdent)
		if err != nil {
			return nil, err
		}
		colType, err := p.expect(TokenIdent)
		if err != nil {
			return nil, err
		}
		stmt.Columns = append(stmt.Columns, &ColumnDef{Name: colName.Literal, Type: colType.Literal})
		
		if p.peek().Type == TokenComma {
			p.next()
		} else {
			break
		}
	}
	p.expect(TokenRParen)
	if p.peek().Type == TokenSemi {
		p.next()
	}

	return stmt, nil
}

func (p *Parser) parseInsert() (*InsertStatement, error) {
	p.expect(TokenInsert)
	p.expect(TokenInto)
	nameTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}
	stmt := &InsertStatement{TableName: nameTok.Literal}

	if p.peek().Type == TokenLParen {
		p.next()
		for {
			col, err := p.expect(TokenIdent)
			if err != nil {
				return nil, err
			}
			stmt.Columns = append(stmt.Columns, col.Literal)
			if p.peek().Type == TokenComma {
				p.next()
			} else {
				break
			}
		}
		p.expect(TokenRParen)
	}

	p.expect(TokenValues)
	
	// parse list of tuples
	for {
		p.expect(TokenLParen)
		var tuple []Expr
		for {
			expr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			tuple = append(tuple, expr)
			if p.peek().Type == TokenComma {
				p.next()
			} else {
				break
			}
		}
		p.expect(TokenRParen)
		stmt.Values = append(stmt.Values, tuple)
		if p.peek().Type == TokenComma {
			p.next()
		} else {
			break
		}
	}

	if p.peek().Type == TokenSemi {
		p.next()
	}

	return stmt, nil
}

// Pratt parsing approach (simplified)
func (p *Parser) parseExpr() (Expr, error) {
	return p.parseBinaryExpr(0)
}

func (p *Parser) parseBinaryExpr(precedence int) (Expr, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.peek()
		if tok.Type == TokenEOF || tok.Type == TokenSemi || tok.Type == TokenComma || tok.Type == TokenRParen || tok.Type == TokenFrom || tok.Type == TokenJoin || tok.Type == TokenWhere || tok.Type == TokenOrder || tok.Type == TokenLimit || tok.Type == TokenOn {
			break
		}

		// simplified precedence
		op := tok.Literal
		if tok.Type == TokenAssign || tok.Type == TokenGT || tok.Type == TokenLT || tok.Type == TokenAnd {
			p.next()
			right, err := p.parsePrimary()
			if err != nil {
				return nil, err
			}
			left = &BinaryExpr{Left: left, Operator: op, Right: right}
		} else {
			break
		}
	}
	return left, nil
}

func (p *Parser) parsePrimary() (Expr, error) {
	tok := p.next()
	switch tok.Type {
	case TokenIdent:
		// Check for ColumnIdent
		if p.peek().Type == TokenDot {
			p.next()
			colTok, err := p.expect(TokenIdent)
			if err != nil {
				return nil, err
			}
			return &ColumnIdent{Table: tok.Literal, Column: colTok.Literal}, nil
		}
		return &Identifier{Value: tok.Literal}, nil
	case "?":
		return &Identifier{Value: "?"}, nil // treating ? as an identifier
	case TokenInt:
		val, _ := strconv.ParseInt(tok.Literal, 10, 64)
		return &IntegerLiteral{Value: val}, nil
	case TokenFloat:
		val, _ := strconv.ParseFloat(tok.Literal, 64)
		return &FloatLiteral{Value: val}, nil
	case TokenString:
		return &StringLiteral{Value: tok.Literal}, nil
	case TokenVectorSim:
		p.expect(TokenLParen)
		vec1, _ := p.parseExpr()
		p.expect(TokenComma)
		vec2, _ := p.parseExpr()
		p.expect(TokenRParen)
		return &VectorSimilarityExpr{Vector1: vec1, Vector2: vec2}, nil
	case TokenVectorDist:
		p.expect(TokenLParen)
		vec1, _ := p.parseExpr()
		p.expect(TokenComma)
		vec2Tok := p.next()
		p.expect(TokenRParen)
		return &VectorDistanceExpr{Vector: vec1, Query: vec2Tok.Literal}, nil
	default:
		return nil, fmt.Errorf("unexpected primary token %v", tok)
	}
}
