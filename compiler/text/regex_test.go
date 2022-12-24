package text

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestTokenization(t *testing.T) {
	e := `[a]\\b+|c`
	l := regexProg.Tokenize([]byte(e))

	ts := []token{
		charsetOpen{},
		char{of: 'a'},
		charsetClose{},
		slash{of: '\\'},
		char{of: 'b'},
		quantity{of: '+'},
		bar{},
		char{of: 'c'},
	}

	for _, tok := range ts {
		assert.True(t, l.Next())
		assert.Equal(t, tok, l.This())
	}
	assert.False(t, l.Next())
}

func TestRegexCompilation(t *testing.T) {
	type testTok struct {
		text string
	}

	p, err := NewLexer(Regex(`(abc*)+`, func(start int, text string) testTok {
		return testTok{text: text}
	}))

	assert.Nil(t, err)

	l := p.Tokenize([]byte("ababcc"))

	assert.True(t, l.Next())
	assert.Equal(t, testTok{"ababcc"}, l.This())
}

func TestParse(t *testing.T) {
	for _, test := range []struct {
		name string
		in   string
		out  expr
	}{
		{
			name: "Any",
			in:   `.`,
			out:  match{start: 0, end: unicode.MaxRune},
		},
		{
			name: "Char",
			in:   `a`,
			out:  match{start: 'a', end: 'a'},
		},
		{
			name: "Esc",
			in:   `\(`,
			out:  match{start: '(', end: '('},
		},
		{
			name: "Option",
			in:   `a?`,
			out: nest{e: choice{
				left:  match{start: 'a', end: 'a'},
				right: empty{},
			}},
		},
		{
			name: "Option",
			in:   `a+`,
			out:  repeat{e: match{start: 'a', end: 'a'}},
		},
		{
			name: "Option",
			in:   `a*`,
			out: nest{choice{
				left:  repeat{e: match{start: 'a', end: 'a'}},
				right: empty{},
			}},
		},
		{
			name: "Seq",
			in:   `ab`,
			out: seq{
				left:  match{start: 'a', end: 'a'},
				right: match{start: 'b', end: 'b'},
			},
		},
		{
			name: "LongSeq",
			in:   `abc`,
			out: seq{
				left: match{start: 'a', end: 'a'},
				right: seq{
					left:  match{start: 'b', end: 'b'},
					right: match{start: 'c', end: 'c'},
				},
			},
		},
		{
			name: "Choice",
			in:   `a|b`,
			out: choice{
				left:  match{start: 'a', end: 'a'},
				right: match{start: 'b', end: 'b'},
			},
		},
		{
			name: "QSeq",
			in:   `ab+`,
			out: seq{
				left:  match{start: 'a', end: 'a'},
				right: repeat{e: match{start: 'b', end: 'b'}},
			},
		},
		{
			name: "Group",
			in:   `(ab)+`,
			out: repeat{nest{e: seq{
				left:  match{start: 'a', end: 'a'},
				right: match{start: 'b', end: 'b'},
			}}},
		},
		{
			name: "CharsetChar",
			in:   `[a]`,
			out:  nest{match{start: 'a', end: 'a'}},
		},
		{
			name: "CharsetChoice",
			in:   `[ab]`,
			out: nest{choice{
				left:  match{start: 'a', end: 'a'},
				right: match{start: 'b', end: 'b'},
			}},
		},
		{
			name: "CharsetRange",
			in:   `[a-z]`,
			out:  nest{match{start: 'a', end: 'z'}},
		},
		{
			name: "InverseCharset",
			in:   `[^b-y]`,
			out: nest{choice{
				left:  match{start: 0, end: 'a'},
				right: match{start: 'z', end: unicode.MaxRune},
			}},
		},
		{
			name: "InverseCharset",
			in:   `[^bcd]`,
			out: nest{choice{
				left:  match{start: 0, end: 'a'},
				right: match{start: 'e', end: unicode.MaxRune},
			}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			toks := regexProg.Tokenize([]byte(test.in))
			expr, err := Parse[token, expr](new(regexRules), toks)
			assert.NoError(t, err)
			assert.Equal(t, test.out, expr)
		})
	}
}
