package parser

import (
	"strconv"
	"strings"

	"github.com/bobappleyard/cezanne/commands/compile/text"
	"github.com/bobappleyard/cezanne/util/must"
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
type importKeyword struct{}
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
func (importKeyword) tok()  {}
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

// This implements stream.Iter
func mapKeywords(t token) token {
	if tok, ok := t.(ident); ok {
		switch tok.name {
		case "import":
			return importKeyword{}
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
	return t
}

func isNewline(scope *[]bool, t token) bool {
	if len(*scope) > 0 && !(*scope)[len(*scope)-1] {
		return false
	}
	tok, ok := t.(whitespace)
	return ok && strings.Contains(tok.text, "\n")
}

func isIgnored(scope *[]bool, t token) bool {
	switch t.(type) {
	case groupOpen:
		*scope = append(*scope, false)
	case blockOpen:
		*scope = append(*scope, true)
	case groupClose, blockClose:
		if len(*scope) > 0 {
			*scope = (*scope)[:len(*scope)-1]
		}
	case comment, whitespace:
		return true
	}
	return false
}

func tokenize(src []byte) ([]token, error) {
	toks, err := lexicon.Tokenize(src)
	if err != nil {
		return nil, err
	}
	var res []token
	var scope []bool
	for _, t := range toks {
		if isNewline(&scope, t) {
			res = append(res, newline{})
			continue
		}
		if isIgnored(&scope, t) {
			continue
		}
		res = append(res, mapKeywords(t))
	}
	return res, nil
}
