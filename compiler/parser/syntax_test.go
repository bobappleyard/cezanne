package parser

import (
	"testing"

	ast "github.com/bobappleyard/cezanne/compiler/ast/shell"
	"github.com/stretchr/testify/assert"
)

func TestParseArgs(t *testing.T) {
	s := Tokenize([]byte(`(a, b, c)`))
	p := parser{}
	args, err := p.parseArgList(s)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, args)
}

func TestParseExpr(t *testing.T) {
	s := Tokenize([]byte(`func(x) = x`))
	p := parser{}
	e, err := p.ParseExpr(s, 0)
	assert.Nil(t, err)
	assert.Equal(t, ast.Lambda{
		Args: []string{"x"},
		Body: []ast.Stmt{
			ast.Return{Value: ast.VarRef{Name: "x"}},
		},
	}, e)
}
