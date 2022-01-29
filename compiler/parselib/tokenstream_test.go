package parselib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLexer(t *testing.T) {
	p := &LexerProgram{
		closeTransitions: []closeTransition{
			{Given: 1, Then: 2},
			{Given: 3, Then: 2},
			{Given: 3, Then: 4},
			{Given: 0, Then: 5},
			{Given: 6, Then: 5},
			{Given: 6, Then: 7},
			{Given: 0, Then: 8},
			{Given: 9, Then: 8},
			{Given: 9, Then: 10},
			{Given: 11, Then: 12},
			{Given: 13, Then: 12},
			{Given: 13, Then: 14},
		},
		moveTransitions: []moveTransition{
			{Given: 0, Min: 'a', Max: 'z', Then: 1},
			{Given: 2, Min: 'a', Max: 'z', Then: 3},
			{Given: 2, Min: '0', Max: '9', Then: 3},
			{Given: 5, Min: '0', Max: '9', Then: 6},
			{Given: 8, Min: '0', Max: '9', Then: 9},
			{Given: 10, Min: '.', Max: '.', Then: 11},
			{Given: 12, Min: '0', Max: '9', Then: 13},
			{Given: 0, Min: '.', Max: '.', Then: 15},
		},
		finalStates: []finalState{
			{Given: 4, TokenID: 1},
			{Given: 7, TokenID: 2},
			{Given: 14, TokenID: 3},
			{Given: 15, TokenID: 4},
		},
		maxState: 16,
	}

	for _, test := range []struct {
		name string
		in   string
		out  []Token
	}{
		{
			name: "Identifier",
			in:   "hello",
			out:  []Token{{ID: 1, End: 5}},
		},
		{
			name: "Integer",
			in:   "123",
			out:  []Token{{ID: 2, End: 3}},
		},
		{
			name: "Float",
			in:   "123.4",
			out:  []Token{{ID: 3, End: 5}},
		},
		{
			name: "IntDot",
			in:   "123.up",
			out: []Token{
				{ID: 2, Start: 0, End: 3},
				{ID: 4, Start: 3, End: 4},
				{ID: 1, Start: 4, End: 6},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			l := p.Lexer([]byte(test.in))
			for _, tok := range test.out {
				require.True(t, l.Next())
				require.Equal(t, tok, l.This())
			}
			require.False(t, l.Next())
		})
	}
}

func TestLexerBuild(t *testing.T) {
	var lp LexerProgram

	s1 := lp.State()
	s2 := lp.State()
	s3 := lp.State()

	end := lp.State()

	lp.Final(end, 1)

	lp.Empty(s1, s2)
	lp.Empty(s2, s3)
	lp.Empty(0, s1)

	lp.Rune(s3, end, '0')

	l := lp.Lexer([]byte("0"))

	assert.True(t, l.Next())
	assert.Equal(t, TokenID(1), l.This().ID)

}
