package parser

import (
	"strconv"
	"strings"

	"github.com/bobappleyard/cezanne/compiler/must"
	"github.com/bobappleyard/cezanne/compiler/stream"
	"github.com/bobappleyard/cezanne/compiler/text"
)

type token interface {
	tok()
}

type comment struct{ text string }
type whitespace struct{ text string }
type newline struct{}
type ident struct{ name string }
type strLit struct{ text string }
type intLit struct{ val int }
type op struct{ of string }
type comma struct{}
type dot struct{}
type groupOpen struct{}
type groupClose struct{}
type blockOpen struct{}
type blockClose struct{}
type funcKeyword struct{}
type objectKeyword struct{}
type effectKeyword struct{}
type varKeyword struct{}
type triggerKeyword struct{}
type handleKeyword struct{}

func (comment) tok()        {}
func (whitespace) tok()     {}
func (newline) tok()        {}
func (ident) tok()          {}
func (strLit) tok()         {}
func (intLit) tok()         {}
func (op) tok()             {}
func (comma) tok()          {}
func (dot) tok()            {}
func (groupOpen) tok()      {}
func (groupClose) tok()     {}
func (blockOpen) tok()      {}
func (blockClose) tok()     {}
func (funcKeyword) tok()    {}
func (objectKeyword) tok()  {}
func (effectKeyword) tok()  {}
func (varKeyword) tok()     {}
func (triggerKeyword) tok() {}
func (handleKeyword) tok()  {}

var lexicon = must.Be(text.NewLexer(
	text.Regex(`//[^\n]*`, func(start int, text string) token {
		return comment{text}
	}),
	text.Regex(`\s+`, func(start int, text string) token {
		return whitespace{text}
	}),
	text.Regex(`\c\w*`, func(start int, text string) token {
		return ident{text}
	}),
	text.Regex(`"([^"]|\\.)*"`, func(start int, text string) token {
		inner, _ := strconv.Unquote(text)
		return strLit{inner}
	}),
	text.Regex(`\d+`, func(start int, text string) token {
		x, _ := strconv.Atoi(text)
		return intLit{x}
	}),
	text.Regex(`-|[+*/><=]+`, func(start int, text string) token {
		return op{text}
	}),
	text.Regex(`,`, func(start int, text string) token {
		return comma{}
	}),
	text.Regex(`\.`, func(start int, text string) token {
		return dot{}
	}),
	text.Regex(`\(`, func(start int, text string) token {
		return groupOpen{}
	}),
	text.Regex(`\)`, func(start int, text string) token {
		return groupClose{}
	}),
	text.Regex(`\{`, func(start int, text string) token {
		return blockOpen{}
	}),
	text.Regex(`\}`, func(start int, text string) token {
		return blockClose{}
	}),
))

type cezanneTokenizer struct {
	src   stream.Iter[token]
	depth int
}

// Err implements stream.Iter
func (t *cezanneTokenizer) Err() error {
	return t.src.Err()
}

// Next implements stream.Iter
func (t *cezanneTokenizer) Next() bool {
	if !t.src.Next() {
		return false
	}
	if t.isIgnored() {
		return t.Next()
	}
	return true
}

// This implements stream.Iter
func (t *cezanneTokenizer) This() token {
	if t.isNewline() {
		return newline{}
	}
	if tok, ok := t.src.This().(ident); ok {
		switch tok.name {
		case "func":
			return funcKeyword{}
		case "effect":
			return effectKeyword{}
		case "var":
			return varKeyword{}
		case "trigger":
			return triggerKeyword{}
		case "handle":
			return handleKeyword{}
		case "object":
			return objectKeyword{}
		}
	}
	return t.src.This()
}

func (t *cezanneTokenizer) isNewline() bool {
	if t.depth > 0 {
		return false
	}
	tok, ok := t.src.This().(whitespace)
	return ok && strings.Contains(tok.text, "\n")
}

func (t *cezanneTokenizer) isIgnored() bool {
	switch t.src.This().(type) {
	case groupOpen:
		t.depth++
	case groupClose:
		t.depth--
	case comment:
		return true
	case whitespace:
		return !t.isNewline()
	}
	return false
}

func tokenize(src []byte) stream.Iter[token] {
	return &cezanneTokenizer{src: lexicon.Tokenize(src)}
}
