package regex

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"unicode"
	"unicode/utf8"

	"github.com/bobappleyard/cezanne/compiler/parselib"
)

var ErrUnexpected = errors.New("unexpected token")

type exprType int

const (
	charExpr exprType = iota
	seqExpr
	choiceExpr
	repeatExpr
)

type expr struct {
	typ      exprType
	min, max rune
	sub      []expr
}

type parser struct {
	escMap  map[rune]expr
	infix   parselib.InfixParser[expr]
	charset parselib.InfixParser[expr]
}

type infix struct {
	prec int
	impl func(s parselib.TokenStream, x expr) (expr, error)
}

// Prec implements parselib.InfixParserDef
func (ip *infix) Prec(t parselib.Token) int {
	return ip.prec
}

// Infix implements parselib.InfixParserDef
func (ip *infix) Infix(s parselib.TokenStream, x expr) (expr, error) {
	return ip.impl(s, x)
}

func (p *parser) init() {
	p.escMap = map[rune]expr{
		'n': {typ: charExpr, min: '\n', max: '\n'},
		't': {typ: charExpr, min: '\t', max: '\t'},
		's': {typ: choiceExpr, sub: []expr{
			{typ: charExpr, min: '\n', max: '\n'},
			{typ: charExpr, min: '\t', max: '\t'},
			{typ: charExpr, min: ' ', max: ' '},
		}},
		'd': {typ: charExpr, min: '0', max: '9'},
		'l': {typ: choiceExpr, sub: []expr{
			{typ: charExpr, min: 'a', max: 'z'},
			{typ: charExpr, min: 'A', max: 'Z'},
		}},
	}

	p.infix.Define(starToken, &infix{3, p.star})
	p.infix.Define(optionToken, &infix{3, p.option})
	p.infix.Define(repeatToken, &infix{3, p.repeat})

	p.infix.Define(inverseToken, &infix{2, p.seq})
	p.infix.Define(chOpToken, &infix{2, p.seq})
	p.infix.Define(grOpToken, &infix{2, p.seq})
	p.infix.Define(escToken, &infix{2, p.seq})
	p.infix.Define(charToken, &infix{2, p.seq})
	p.infix.Define(rangeToken, &infix{2, p.seq})
	p.infix.Define(anyToken, &infix{2, p.seq})

	p.infix.Define(choiceToken, &infix{1, p.choice})

	p.charset.Define(rangeToken, &infix{2, p.charRange})

	p.charset.Define(inverseToken, &infix{1, p.singleChar})
	p.charset.Define(grOpToken, &infix{1, p.singleChar})
	p.charset.Define(grClToken, &infix{1, p.singleChar})
	p.charset.Define(starToken, &infix{1, p.singleChar})
	p.charset.Define(optionToken, &infix{1, p.singleChar})
	p.charset.Define(repeatToken, &infix{1, p.singleChar})
	p.charset.Define(choiceToken, &infix{1, p.singleChar})
	p.charset.Define(anyToken, &infix{1, p.singleChar})
	p.charset.Define(escToken, &infix{1, p.singleChar})
	p.charset.Define(charToken, &infix{1, p.singleChar})
}

func (p *parser) parse(s parselib.TokenStream, prec int) (expr, error) {
	if !s.Next() {
		return expr{typ: seqExpr}, s.Err()
	}
	var left expr
	t := s.This()

	switch t.ID {
	case charToken, escToken, rangeToken:
		left = p.charExpr(s, t)

	case anyToken:
		left = expr{typ: charExpr, min: 0, max: unicode.MaxRune}

	case grOpToken:
		e, err := p.parseInner(s)
		if err != nil {
			return expr{}, err
		}
		left = e

	case grClToken:
		s.Back()
		return expr{typ: seqExpr}, nil

	case choiceToken:
		s.Back()
		left = expr{typ: seqExpr}

	case chOpToken:
		e, err := p.parseCharsetExpr(s)
		if err != nil {
			return expr{}, err
		}
		left = e

	default:
		return expr{}, p.unexpected(s, t)

	}

	return p.infix.Parse(s, prec, left)
}

func (p *parser) seq(s parselib.TokenStream, x expr) (expr, error) {
	s.Back()

	y, err := p.parse(s, 2)
	if err != nil {
		return expr{}, err
	}

	return expr{
		typ: seqExpr,
		sub: []expr{x, y},
	}, nil
}

func (p *parser) choice(s parselib.TokenStream, x expr) (expr, error) {
	y, err := p.parse(s, 1)
	if err != nil {
		return expr{}, err
	}

	return expr{
		typ: choiceExpr,
		sub: []expr{x, y},
	}, nil
}

func (p *parser) star(s parselib.TokenStream, x expr) (expr, error) {
	return expr{
		typ: choiceExpr,
		sub: []expr{
			{typ: seqExpr},
			{typ: repeatExpr, sub: []expr{x}},
		},
	}, nil
}

func (p *parser) option(s parselib.TokenStream, x expr) (expr, error) {
	return expr{
		typ: choiceExpr,
		sub: []expr{{typ: seqExpr}, x},
	}, nil
}

func (p *parser) repeat(s parselib.TokenStream, x expr) (expr, error) {
	return expr{typ: repeatExpr, sub: []expr{x}}, nil
}

func (p *parser) parseInner(s parselib.TokenStream) (expr, error) {
	e, err := p.parse(s, 0)
	if err != nil {
		return expr{}, err
	}
	return p.expectSuffix(s, grClToken, e)
}

func (p *parser) parseCharsetExpr(s parselib.TokenStream) (expr, error) {
	if !s.Next() {
		return expr{}, p.lexError(s)
	}
	inverse := s.This().ID == inverseToken
	if !inverse {
		s.Back()
	}
	e, err := p.parseCharset(s)
	if err != nil {
		return expr{}, err
	}
	if inverse {
		e = invertCharset(e)
	}
	return p.expectSuffix(s, chClToken, e)
}

func (p *parser) expectSuffix(s parselib.TokenStream, tok parselib.TokenID, e expr) (expr, error) {
	if !s.Next() {
		return expr{}, p.lexError(s)
	}
	if s.This().ID != tok {
		return expr{}, p.unexpected(s, s.This())
	}

	return e, nil
}

func (p *parser) parseCharset(s parselib.TokenStream) (expr, error) {
	if !s.Next() {
		return expr{}, p.lexError(s)
	}

	t := s.This()

	switch t.ID {
	case charToken, escToken, grOpToken, grClToken, optionToken, choiceToken, starToken, repeatToken:
		return p.charset.Parse(s, 0, p.charExpr(s, t))

	default:
		return expr{}, p.unexpected(s, t)
	}
}

func (p *parser) singleChar(s parselib.TokenStream, x expr) (expr, error) {
	s.Back()
	y, err := p.parseCharset(s)
	if err != nil {
		return expr{}, err
	}
	return expr{typ: choiceExpr, sub: []expr{x, y}}, nil
}

func (p *parser) charRange(s parselib.TokenStream, x expr) (expr, error) {
	if x.min != x.max {
		return expr{}, fmt.Errorf("%q: %w", "-", ErrUnexpected)
	}
	if !s.Next() {
		return expr{}, p.lexError(s)
	}
	t := s.This()
	if t.ID == rangeToken {
		return expr{}, p.unexpected(s, t)
	}

	return expr{typ: charExpr, min: x.min, max: tokRune(s, t)}, nil
}

func (p *parser) charExpr(s parselib.TokenStream, t parselib.Token) expr {
	c := tokRune(s, t)
	if e, ok := p.escMap[c]; t.ID == escToken && ok {
		return e
	}
	return expr{typ: charExpr, min: c, max: c}
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

func invertCharset(e expr) expr {
	chars := findChars(e, nil)
	sort.Slice(chars, func(i, j int) bool {
		return chars[i].min < chars[j].min
	})

	var last rune
	var res []expr
	for _, c := range chars {
		if c.min <= last {
			continue
		}
		res = append(res, expr{typ: charExpr, min: last, max: c.min - 1})
		last = c.max + 1
	}
	res = append(res, expr{typ: charExpr, min: last, max: unicode.MaxRune})

	return expr{typ: choiceExpr, sub: res}
}

func findChars(e expr, res []expr) []expr {
	switch e.typ {
	case charExpr:
		return append(res, e)
	case choiceExpr:
		for _, e := range e.sub {
			res = findChars(e, res)
		}
		return res
	}
	panic("unreachable")
}

func tokRune(s parselib.TokenStream, t parselib.Token) rune {
	if t.ID == escToken {
		c, _ := utf8.DecodeRuneInString(s.Text(t)[1:])
		return c
	}
	c, _ := utf8.DecodeRuneInString(s.Text(t))
	return c
}
