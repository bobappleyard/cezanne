package backend

import (
	"fmt"
	"testing"

	"github.com/bobappleyard/cezanne/commands/compile/ast"
	"github.com/bobappleyard/cezanne/format/symtab"
)

func TestInterpretExpr(t *testing.T) {
	var m method

	var syms symtab.Symtab

	s := scope{syms: &syms}

	interpretExpr(s, &m, ast.Invoke{
		Object: ast.Create{Methods: []ast.Method{
			{
				Name: syms.SymbolID("f"),
				Args: []symtab.Symbol{syms.SymbolID("x")},
				Body: ast.Create{Methods: []ast.Method{
					{
						Name: syms.SymbolID("g"),
						Args: []symtab.Symbol{syms.SymbolID("y")},
						Body: ast.Ref{Name: syms.SymbolID("x")},
					},
				}},
			},
		}},
		Name: syms.SymbolID("f"),
		Args: []ast.Expr{ast.Int{Value: 1}},
	})

	for _, s := range m.steps {
		fmt.Printf("%#v\n", s)
	}
	// t.Fail()
}
