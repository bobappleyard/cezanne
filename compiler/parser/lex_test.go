package parser

import (
	"testing"

	"github.com/bobappleyard/cezanne/compiler/parselib"
	"github.com/stretchr/testify/assert"
)

func TestLex(t *testing.T) {
	s := Tokenize([]byte(`

import math

@record(.x, .y)
	`))

	toks := []parselib.Token{
		{ID: newlineTok, Start: 0, End: 2},
		{ID: idTok, Start: 2, End: 8},
		{ID: idTok, Start: 9, End: 13},
		{ID: newlineTok, Start: 13, End: 15},
		{ID: atTok, Start: 15, End: 16},
		{ID: idTok, Start: 16, End: 22},
		{ID: groupStartTok, Start: 22, End: 23},
		{ID: dotTok, Start: 23, End: 24},
		{ID: idTok, Start: 24, End: 25},
		{ID: commaTok, Start: 25, End: 26},
		{ID: dotTok, Start: 27, End: 28},
		{ID: idTok, Start: 28, End: 29},
		{ID: groupEndTok, Start: 29, End: 30},
		{ID: newlineTok, Start: 30, End: 32},
	}

	for _, tok := range toks {
		assert.True(t, s.Next())
		assert.Equal(t, tok, s.This())
	}
	assert.False(t, s.Next())
	assert.Nil(t, s.Err())
}
