package parser

import (
	"strings"

	"github.com/bobappleyard/cezanne/compiler/parselib"
	"github.com/bobappleyard/cezanne/compiler/regex"
)

var lex parselib.LexerProgram

func Tokenize(src []byte) parselib.TokenStream {
	return parselib.LL1(&tokenStream{base: lex.Lexer(src)})
}

type tokenStream struct {
	base  parselib.TokenSource
	depth int
}

// Err implements parselib.TokenStream
func (s *tokenStream) Err() error {
	return s.base.Err()
}

// Next implements parselib.TokenStream
func (s *tokenStream) Next() bool {
	if !s.base.Next() {
		return false
	}
	t := s.base.This()

	switch t.ID {
	case spaceTok:
		if s.depth == 0 && strings.Contains(s.base.Text(t), "\n") {
			return true
		}
		return s.Next()
	case groupStartTok:
		s.depth++
	case groupEndTok:
		s.depth--
	}
	return true
}

// Text implements parselib.TokenStream
func (s *tokenStream) Text(t parselib.Token) string {
	return s.base.Text(t)
}

// This implements parselib.TokenStream
func (s *tokenStream) This() parselib.Token {
	t := s.base.This()
	if t.ID == spaceTok {
		// if it got through the filter in Next this must be true
		t.ID = newlineTok
	}
	return t
}

const (
	invalidTok = iota
	newlineTok
	spaceTok
	idTok
	intTok
	strTok
	fltTok
	blockStartTok
	blockEndTok
	groupStartTok
	groupEndTok
	dotTok
	atTok
	commaTok
	eqTok
	mulTok
	plusTok
)

func init() {
	for _, def := range []struct {
		tok parselib.TokenID
		re  string
	}{
		{spaceTok, `\s+`},
		{idTok, `[\l_][\l\d_]*`},
		{intTok, `\d+`},
		{strTok, `"([^"]|\\.)*"`},
		{fltTok, `\d+\.\d+`},
		{blockStartTok, `\{`},
		{blockEndTok, `\}`},
		{groupStartTok, `\(`},
		{groupEndTok, `\)`},
		{dotTok, `\.`},
		{atTok, `@`},
		{commaTok, `,`},
		{eqTok, `=`},
		{mulTok, `\*`},
		{plusTok, `\+`},
	} {
		regex.Regex(&lex, def.tok, def.re)
	}

}
