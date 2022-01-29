package regex

import (
	"testing"
	"unicode"

	"github.com/bobappleyard/cezanne/compiler/parselib"
	"github.com/stretchr/testify/assert"
)

func TestSyntax(t *testing.T) {
	for _, test := range []struct {
		name string
		src  string
		res  expr
	}{
		{
			name: "AnyChar",
			src:  ".",
			res:  expr{typ: charExpr, min: 0, max: unicode.MaxRune},
		},
		{
			name: "SingleChar",
			src:  "a",
			res:  expr{typ: charExpr, min: 'a', max: 'a'},
		},
		{
			name: "EscChar",
			src:  "\\\\",
			res:  expr{typ: charExpr, min: '\\', max: '\\'},
		},
		{
			name: "CharPair",
			src:  "ab",
			res: expr{typ: seqExpr, sub: []expr{
				{typ: charExpr, min: 'a', max: 'a'},
				{typ: charExpr, min: 'b', max: 'b'},
			}},
		},
		{
			name: "ChoiceAndSeq",
			src:  "a|bc",
			res: expr{typ: choiceExpr, sub: []expr{
				{typ: charExpr, min: 'a', max: 'a'},
				{typ: seqExpr, sub: []expr{
					{typ: charExpr, min: 'b', max: 'b'},
					{typ: charExpr, min: 'c', max: 'c'},
				}},
			}},
		},
		{
			name: "GroupChoice",
			src:  "(a|b)c",
			res: expr{typ: seqExpr, sub: []expr{
				{typ: choiceExpr, sub: []expr{
					{typ: charExpr, min: 'a', max: 'a'},
					{typ: charExpr, min: 'b', max: 'b'},
				}},
				{typ: charExpr, min: 'c', max: 'c'},
			}},
		},
		{
			name: "GroupChoiceEmpty",
			src:  "(a|)c",
			res: expr{typ: seqExpr, sub: []expr{
				{typ: choiceExpr, sub: []expr{
					{typ: charExpr, min: 'a', max: 'a'},
					{typ: seqExpr},
				}},
				{typ: charExpr, min: 'c', max: 'c'},
			}},
		},
		{
			name: "GroupChoiceEmptyLeft",
			src:  "(|a)c",
			res: expr{typ: seqExpr, sub: []expr{
				{typ: choiceExpr, sub: []expr{
					{typ: seqExpr},
					{typ: charExpr, min: 'a', max: 'a'},
				}},
				{typ: charExpr, min: 'c', max: 'c'},
			}},
		},
		{
			name: "Option",
			src:  "a?",
			res: expr{typ: choiceExpr, sub: []expr{
				{typ: seqExpr},
				{typ: charExpr, min: 'a', max: 'a'},
			}},
		},
		{
			name: "Repeat",
			src:  "a+",
			res: expr{typ: repeatExpr, sub: []expr{
				{typ: charExpr, min: 'a', max: 'a'},
			}},
		},
		{
			name: "Star",
			src:  "a*",
			res: expr{typ: choiceExpr, sub: []expr{
				{typ: seqExpr},
				{typ: repeatExpr, sub: []expr{
					{typ: charExpr, min: 'a', max: 'a'},
				}},
			}},
		},
		{
			name: "CharsetEnum",
			src:  "[abc]",
			res: expr{typ: choiceExpr, sub: []expr{
				{typ: charExpr, min: 'a', max: 'a'},
				{typ: choiceExpr, sub: []expr{
					{typ: charExpr, min: 'b', max: 'b'},
					{typ: charExpr, min: 'c', max: 'c'},
				}},
			}},
		},
		{
			name: "CharsetRange",
			src:  "[a-c]",
			res:  expr{typ: charExpr, min: 'a', max: 'c'},
		},
		{
			name: "CharsetRangeEnum",
			src:  "[a-z0-9_]",
			res: expr{typ: choiceExpr, sub: []expr{
				{typ: charExpr, min: 'a', max: 'z'},
				{typ: choiceExpr, sub: []expr{
					{typ: charExpr, min: '0', max: '9'},
					{typ: charExpr, min: '_', max: '_'},
				}},
			}},
		},
		{
			name: "InverseCharset",
			src:  "[^b]",
			res: expr{typ: choiceExpr, sub: []expr{
				{typ: charExpr, min: 0, max: 'a'},
				{typ: charExpr, min: 'c', max: unicode.MaxRune},
			}},
		},
		{
			name: "InverseCharsetMulti",
			src:  "[^b0-9]",
			res: expr{typ: choiceExpr, sub: []expr{
				{typ: charExpr, min: 0, max: '0' - 1},
				{typ: charExpr, min: '9' + 1, max: 'a'},
				{typ: charExpr, min: 'c', max: unicode.MaxRune},
			}},
		},
		{
			name: "CharsetRepeat",
			src:  "[0-9]+",
			res: expr{typ: repeatExpr, sub: []expr{
				{typ: charExpr, min: '0', max: '9'},
			}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var p parser
			p.init()
			s := regexProg.Lexer([]byte(test.src))
			e, err := p.parse(parselib.LL1(s), 0)
			assert.Nil(t, err)
			assert.Equal(t, test.res, e)
		})
	}
}
