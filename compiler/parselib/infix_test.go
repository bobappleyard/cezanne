package parselib

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	Invalid = iota
	LitTok
	PlusTok
	MulTok
)

var errUnexpectedToken = errors.New("unexpected token")

type ast struct {
	n, v string
	sub  []ast
}

type leftOp struct {
	parser func(s TokenStream, prec int) (ast, error)
	prec   int
	n      string
}

// Infix implements InfixParserDef
func (o *leftOp) Infix(s TokenStream, left ast) (ast, error) {
	right, err := o.parser(s, o.prec)
	if err != nil {
		return ast{}, err
	}

	return ast{
		n:   o.n,
		sub: []ast{left, right},
	}, nil
}

// Prec implements InfixParserDef
func (o *leftOp) Prec(t Token) int {
	return o.prec
}

func parseTestExpr(p InfixParser[ast], s TokenStream, prec int) (ast, error) {
	if !s.Next() {
		return ast{}, s.Err()
	}

	t := s.This()
	if t.ID != LitTok {
		return ast{}, errUnexpectedToken
	}

	left := ast{n: "literal", v: s.Text(t)}

	return p.Parse(s, prec, left)
}

func TestInfixParser(t *testing.T) {
	lit := func(v string) ast { return ast{n: "literal", v: v} }
	op := func(n string, sub ...ast) ast { return ast{n: n, sub: sub} }
	toks := func(ids ...TokenID) []Token {
		res := make([]Token, len(ids))
		for i, id := range ids {
			res[i] = Token{ID: id, Start: i, End: i + 1}
		}
		return res
	}

	var p InfixParser[ast]
	parse := func(s TokenStream, prec int) (ast, error) {
		return parseTestExpr(p, s, prec)
	}
	p.Define(PlusTok, &leftOp{parse, 1, "add"})
	p.Define(MulTok, &leftOp{parse, 2, "mul"})

	for _, test := range []struct {
		name string
		src  string
		toks []Token
		expr ast
	}{
		{
			name: "Literal",
			src:  "1",
			toks: toks(LitTok),
			expr: lit("1"),
		},
		{
			name: "Simple",
			src:  "1+2",
			toks: toks(LitTok, PlusTok, LitTok),
			expr: op("add", lit("1"), lit("2")),
		},
		{
			name: "Left",
			src:  "1+2+3",
			toks: toks(LitTok, PlusTok, LitTok, PlusTok, LitTok),
			expr: op("add", op("add", lit("1"), lit("2")), lit("3")),
		},
		{
			name: "Mix",
			src:  "1+2*3+4",
			toks: toks(LitTok, PlusTok, LitTok, MulTok, LitTok, PlusTok, LitTok),
			expr: op("add", op("add", lit("1"), op("mul", lit("2"), lit("3"))), lit("4")),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := NewMockTokenStream(test.src, test.toks)
			expr, err := parseTestExpr(p, LL1(s), 0)
			require.Nil(t, err)
			require.Equal(t, test.expr, expr)
		})
	}

	t.Run("Fail", func(t *testing.T) {
		s := NewMockTokenStream("1++", toks(LitTok, PlusTok, PlusTok))
		expr, err := parseTestExpr(p, LL1(s), 0)
		require.Equal(t, errUnexpectedToken, err)
		require.Equal(t, ast{}, expr)
	})
}

func TestInfixLexer(t *testing.T) {
	lit := func(v string) ast { return ast{n: "literal", v: v} }
	op := func(n string, sub ...ast) ast { return ast{n: n, sub: sub} }

	var parser InfixParser[ast]
	parse := func(s TokenStream, prec int) (ast, error) {
		return parseTestExpr(parser, s, prec)
	}
	parser.Define(PlusTok, &leftOp{parse, 1, "add"})
	parser.Define(MulTok, &leftOp{parse, 2, "mul"})

	prog := &LexerProgram{
		moveTransitions: []moveTransition{
			{Given: 0, Min: '0', Max: '9', Then: 1},
			{Given: 1, Min: '0', Max: '9', Then: 1},
			{Given: 0, Min: '+', Max: '+', Then: 2},
			{Given: 0, Min: '*', Max: '*', Then: 3},
		},
		finalStates: []finalState{
			{Given: 1, TokenID: LitTok},
			{Given: 2, TokenID: PlusTok},
			{Given: 3, TokenID: MulTok},
		},
		maxState: 4,
	}

	for _, test := range []struct {
		name string
		src  string
		expr ast
	}{
		{
			name: "Literal",
			src:  "1",
			expr: lit("1"),
		},
		{
			name: "Simple",
			src:  "1+2",
			expr: op("add", lit("1"), lit("2")),
		},
		{
			name: "Left",
			src:  "1+2+3",
			expr: op("add", op("add", lit("1"), lit("2")), lit("3")),
		},
		{
			name: "Mix",
			src:  "1+2*3+4",
			expr: op("add", op("add", lit("1"), op("mul", lit("2"), lit("3"))), lit("4")),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := prog.Lexer([]byte(test.src))
			expr, err := parseTestExpr(parser, LL1(s), 0)
			require.Nil(t, err)
			require.Equal(t, test.expr, expr)
		})
	}

}
