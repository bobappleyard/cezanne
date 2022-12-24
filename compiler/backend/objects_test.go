package backend

import (
	"fmt"
	"testing"

	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/sexpr"
	"github.com/stretchr/testify/assert"
)

func TestNamespace(t *testing.T) {
	ns := Namespace{}

	ns.Method("str")
	ns.Method("call")
	ns.Method("hello")

	ns.DefineClass([]Mapping{
		{ns.Method("str"), "1:str"},
		{ns.Method("hello"), "1:hello"},
	})

	ns.DefineClass([]Mapping{
		{ns.Method("str"), "2:str"},
		{ns.Method("hello"), "2:hello"},
	})

	assert.Equal(t, "(vector 0 1:str 0 2:str 2 1:hello 2 2:hello)", fmt.Sprint(ns.Render()))
}

func TestCompile(t *testing.T) {
	ctx := Context{
		module: new(Module),
	}

	expr := ctx.CompileExpr(ast.Create{Methods: []ast.Method{
		{
			Name: "hello",
			Args: []string{"x"},
			Body: ast.Invoke{Object: ast.Ref{Name: "x"}, Name: "print", Args: []ast.Expr{}},
		},
		{
			Name: "field",
			Args: []string{},
			Body: ast.Ref{Name: "y"},
		},
		{
			Name: "method",
			Args: []string{},
			Body: ast.Invoke{
				Object: ast.Ref{Name: "this"},
				Name:   "field",
				Args:   nil,
			},
		},
	}})

	// (vector 0 y)
	// (define (0:hello this x) (method-invoke x 1))
	// (define (0:field this) (vector-ref this 1))
	// (define (0:method this) (method-invoke this 2))
	// (vector 0 0:hello 0 #f 2 0:field 3 0:method)

	module := []sexpr.Node{
		sexpr.DefineFunc(
			"0:hello",
			[]string{"this", "x"},
			sexpr.Call("method-invoke", sexpr.Var("x"), sexpr.Int(0)),
		),
		sexpr.DefineFunc(
			"0:field",
			[]string{"this"},
			sexpr.Call("vector-ref", sexpr.Var("this"), sexpr.Int(1)),
		),
		sexpr.DefineFunc(
			"0:method",
			[]string{"this"},
			sexpr.Call("method-invoke", sexpr.Var("this"), sexpr.Int(2)),
		),
		sexpr.DefineVar("%methods", sexpr.Call("vector",
			sexpr.Int(0), sexpr.Bool(false),
			sexpr.Int(1), sexpr.Var("0:hello"),
			sexpr.Int(2), sexpr.Var("0:field"),
			sexpr.Int(3), sexpr.Var("0:method"),
		)),
	}

	assert.Equal(t, sexpr.Call("vector", sexpr.Int(0), sexpr.Var("y")), expr)
	assert.Equal(t, module, ctx.Render())
}

func TestCompileModule(t *testing.T) {

	ctx := Context{
		module: new(Module),
	}

	ctx.CompileModule(&ast.Module{
		Name: "test",
		Funcs: []ast.Method{{
			Name: "double",
			Args: []string{"x"},
			Body: ast.Invoke{
				Object: ast.Ref{Name: "x"},
				Name:   "add",
				Args:   []ast.Expr{ast.Ref{Name: "x"}},
			},
		}},
	})

	module := []sexpr.Node{
		sexpr.DefineFunc(
			"0:double",
			[]string{"this", "x"},
			sexpr.Call("method-invoke", sexpr.Var("x"), sexpr.Int(0), sexpr.Var("x")),
		),
		sexpr.DefineVar("module:test", sexpr.Call("vector", sexpr.Int(0))),
		sexpr.DefineVar("%methods", sexpr.Call("vector",
			sexpr.Int(0), sexpr.Bool(false),
			sexpr.Int(1), sexpr.Var("0:double"),
		)),
	}

	assert.Equal(t, module, ctx.Render())
}
