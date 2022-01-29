package parser

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	ast "github.com/bobappleyard/cezanne/compiler/ast/shell"
	"github.com/bobappleyard/cezanne/compiler/parselib"
)

var ErrUnexpected = errors.New("unexpected")

type parser struct {
	exprInfix parselib.InfixParser[ast.Expr]
}

func (p *parser) ParseExpr(s parselib.TokenStream, prec int) (ast.Expr, error) {
	if !s.Next() {
		return nil, p.lexError(s)
	}

	t := s.This()
	var left ast.Expr

	switch t.ID {
	case newlineTok:
		e, err := p.ParseExpr(s, prec)
		if err != nil {
			return nil, err
		}
		left = e

	case intTok:
		x, err := strconv.Atoi(s.Text(t))
		if err != nil {
			return nil, err
		}
		left = ast.IntLit{Value: x}

	case strTok:
		x, err := strconv.Unquote(s.Text(t))
		if err != nil {
			return nil, err
		}
		left = ast.StrLit{Value: x}

	case fltTok:
		x, err := strconv.ParseFloat(s.Text(t), 64)
		if err != nil {
			return nil, err
		}
		left = ast.FltLit{Value: x}

	case idTok:
		e, err := p.keywordOrVariable(s, prec, t)
		if err != nil {
			return nil, err
		}
		left = e

	case groupStartTok:
		e, err := p.ParseExpr(s, prec)
		if err != nil {
			return nil, err
		}
		if err := p.expect(s, groupEndTok); err != nil {
			return nil, err
		}
		left = e

	default:
		return nil, p.unexpected(s, t)
	}

	return p.exprInfix.Parse(s, prec, left)
}

func (p *parser) keywordOrVariable(s parselib.TokenStream, prec int, t parselib.Token) (ast.Expr, error) {
	text := s.Text(t)
	switch text {
	case "func":
		return p.parseFunc(s)

	default:
		return ast.VarRef{Name: text}, nil
	}
}

func (p *parser) parseFunc(s parselib.TokenStream) (ast.Expr, error) {
	args, err := p.parseArgList(s)
	if err != nil {
		return nil, err
	}
	if !s.Next() {
		return nil, p.lexError(s)
	}
	if s.This().ID == eqTok {
		val, err := p.ParseExpr(s, 0)
		if err != nil {
			return nil, err
		}
		return ast.Lambda{
			Args: args,
			Body: []ast.Stmt{ast.Return{Value: val}},
		}, nil
	}
	s.Back()
	block, err := p.parseBlock(s)
	if err != nil {
		return nil, err
	}
	return ast.Lambda{
		Args: args,
		Body: block,
	}, nil
}

func (p *parser) parseArgList(s parselib.TokenStream) ([]string, error) {
	if err := p.expect(s, groupStartTok); err != nil {
		return nil, err
	}
	var args []string
	k := false
	for {
		if !s.Next() {
			return nil, p.lexError(s)
		}
		t := s.This()
		switch t.ID {
		case groupEndTok:
			return args, nil

		case idTok:
			if k {
				return nil, p.unexpected(s, t)
			}
			k = true
			args = append(args, s.Text(t))

		case commaTok:
			if !k {
				return nil, p.unexpected(s, t)
			}
			k = false

		default:
			return nil, p.unexpected(s, t)
		}
	}
}

func (p *parser) parseBlock(s parselib.TokenStream) ([]ast.Stmt, error) {
	return nil, nil
}

func (p *parser) expect(s parselib.TokenStream, id parselib.TokenID) error {
	if !s.Next() {
		return p.lexError(s)
	}
	if s.This().ID != id {
		return p.unexpected(s, s.This())
	}
	return nil
}

func (p *parser) lexError(s parselib.TokenStream) error {
	if s.Err() == nil {
		return io.ErrUnexpectedEOF
	}
	return s.Err()
}

func (p *parser) unexpected(s parselib.TokenStream, t parselib.Token) error {
	return fmt.Errorf("%q: %w", s.Text(t), ErrUnexpected)
}

type leftOp struct {
	prec   int
	method string
	parser *parser
}

// Infix implements parselib.InfixParserDef
func (o *leftOp) Infix(s parselib.TokenStream, left ast.Expr) (ast.Expr, error) {
	right, err := o.parser.ParseExpr(s, o.prec)
	if err != nil {
		return nil, err
	}

	return ast.Call{
		Fn: ast.MemberRef{
			Object: left,
			Name:   o.method,
		},
		Args: []ast.Expr{right},
	}, nil
}

// Prec implements parselib.InfixParserDef
func (o *leftOp) Prec(t parselib.Token) int {
	return o.prec
}

var _ parselib.InfixParserDef[ast.Expr] = new(leftOp)
