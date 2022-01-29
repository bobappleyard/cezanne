package regex

import (
	"testing"

	"github.com/bobappleyard/cezanne/compiler/parselib"
	"github.com/stretchr/testify/assert"
)

func TestTokenization(t *testing.T) {
	e := `[a]\\b+|c`
	l := regexProg.Lexer([]byte(e))

	ts := []parselib.Token{
		{ID: chOpToken, Start: 0, End: 1},
		{ID: charToken, Start: 1, End: 2},
		{ID: chClToken, Start: 2, End: 3},
		{ID: escToken, Start: 3, End: 5},
		{ID: charToken, Start: 5, End: 6},
		{ID: repeatToken, Start: 6, End: 7},
		{ID: choiceToken, Start: 7, End: 8},
		{ID: charToken, Start: 8, End: 9},
	}

	for _, tok := range ts {
		assert.True(t, l.Next())
		assert.Equal(t, tok, l.This())
	}
	assert.False(t, l.Next())
}
