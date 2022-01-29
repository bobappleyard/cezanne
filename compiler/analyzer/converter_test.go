package analyzer

import (
	"fmt"
	"testing"

	"github.com/bobappleyard/cezanne/compiler/ast/core"
	"github.com/bobappleyard/cezanne/compiler/ast/shell"
)

func TestConversion(t *testing.T) {

	block := []shell.Stmt{
		shell.VarDef{Name: "x", Value: shell.IntLit{Value: 1}},
		shell.If{
			Cond: shell.Call{
				Fn: shell.MemberRef{
					Object: shell.VarRef{Name: "x"},
					Name:   "eq",
				},
				Args: []shell.Expr{
					shell.IntLit{Value: 2},
				},
			},
			Then: []shell.Stmt{
				shell.VarDef{Name: "x", Value: shell.IntLit{Value: 4}},
				shell.Return{Value: shell.Call{
					Fn: shell.MemberRef{
						Object: shell.VarRef{Name: "x"},
						Name:   "to",
					},
					Args: []shell.Expr{
						shell.VarRef{Name: "y"},
						shell.VarRef{Name: "x"},
					},
				}},
			},
		},
	}

	c := &converter{
		unit:  new(core.Unit),
		block: new(core.Block),
		args:  []string{"y"},
	}

	c.convertBlock(block)

	t.Log(c.err)

	for i, s := range c.block.Body {
		fmt.Printf("%3d: %#v\n", i, s)
	}

	// t.Fail()
}
