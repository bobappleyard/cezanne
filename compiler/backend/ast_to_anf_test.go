package backend

import (
	"fmt"
	"testing"

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

	for _, s := range m.steps {
		fmt.Printf("%#v\n", s)
	}
	// t.Fail()
}
