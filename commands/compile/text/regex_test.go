package text

import (
	"testing"
	"unicode"

	"github.com/bobappleyard/cezanne/util/assert"
)

func TestTokenization(t *testing.T) {
	e := `[a]\\b+|c`
	l, err := regexProg.Tokenize([]byte(e))
	assert.Nil(t, err)

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

	assert.Equal(t, l, ts)
}

func TestRegexCompilation(t *testing.T) {
	type testTok struct {
		text string
	}

	p, err := NewLexer(Regex(`d(abc*)+`, func(start int, text string) testTok {
		return testTok{text: text}
	}))

	assert.Nil(t, err)

	l, err := p.Tokenize([]byte("dababccdab"))

	assert.Nil(t, err)
	assert.Equal(t, l, []testTok{{"dababcc"}, {"dab"}})
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
			name: "Named",
			in:   `\n`,
			out:  nest{match{start: '\n', end: '\n'}},
		},
		{
			name: "Option",
			in:   `a?`,
			out: nest{nested: choice{
				left:  match{start: 'a', end: 'a'},
				right: empty{},
			}},
		},
		{
			name: "Option",
			in:   `a+`,
			out:  repeat{repeated: match{start: 'a', end: 'a'}},
		},
		{
			name: "Option",
			in:   `a*`,
			out: nest{choice{
				left:  repeat{repeated: match{start: 'a', end: 'a'}},
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
				right: repeat{repeated: match{start: 'b', end: 'b'}},
			},
		},
		{
			name: "Group",
			in:   `(ab)+`,
			out: repeat{nest{nested: seq{
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
			name: "CharsetMetas",
			in:   `[|+.]`,
			out: nest{choice{
				left: choice{
					left:  match{start: '|', end: '|'},
					right: match{start: '+', end: '+'},
				},
				right: match{start: '.', end: '.'}},
			},
		},
		{
			name: "CharsetNamed",
			in:   `[\n]`,
			out:  nest{match{start: '\n', end: '\n'}},
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
			toks, err := regexProg.Tokenize([]byte(test.in))
			assert.Nil(t, err)
			expr, err := Parse[token, expr](regexGrammar, toks)
			assert.Nil(t, err)
			assert.Equal(t, expr, test.out)
		})
	}
}
