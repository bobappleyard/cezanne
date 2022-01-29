package regex

import (
	"testing"

	"github.com/bobappleyard/cezanne/compiler/parselib"
	"github.com/stretchr/testify/assert"
)

func TestRegexCompilation(t *testing.T) {

	var p parselib.LexerProgram

	err := Regex(&p, 1, "(abc*)+")

	assert.Nil(t, err)

	l := p.Lexer([]byte("ababcc"))

	assert.True(t, l.Next())
	assert.Equal(t, parselib.Token{ID: 1, Start: 0, End: 6}, l.This())
}
