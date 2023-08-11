package backend

import (
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/compiler/ast"
)

func TestInterpretExpr(t *testing.T) {
	var m method
	var s scope

	interpretExpr(s, &m, ast.Invoke{
		Object: ast.Create{Methods: []ast.Method{
			{
				Name: "f",
				Args: []string{"x"},
				Body: ast.Create{Methods: []ast.Method{
					{
						Name: "g",
						Args: []string{"y"},
						Body: ast.Ref{Name: "x"},
					},
				}},
			},
		}},
		Name: "f",
		Args: []ast.Expr{ast.Int{Value: 1}},
	})

	// t.Fail()
}

func TestReuseVariables(t *testing.T) {
	m := method{
		steps: []step{
			intStep{into: 0},
			localStep{from: 0, into: 1},
			createStep{into: 3, fields: []variable{0, 1}},
			fieldStep{from: 3, into: 4},
			callMethodStep{into: 5, object: 4, params: []variable{3}},
		},
		varc: 5,
	}

	reuseVariables(&m, false)

	assert.Equal(t, m.steps, []step{
		intStep{into: 0},
		localStep{from: 0, into: 1},
		createStep{into: 0, fields: []variable{0, 1}},
		fieldStep{from: 0, into: 1},
		callMethodStep{into: 0, object: 1, params: []variable{0}},
	})
	assert.Equal(t, m.varc, 2)
}
