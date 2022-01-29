package analyzer

import (
	"testing"

	"github.com/bobappleyard/cezanne/compiler/ast/shell"
	"github.com/stretchr/testify/assert"
)

func TestRemoveClosures(t *testing.T) {
	closure := shell.MemberRef{
		Object: shell.MemberRef{
			Object: shell.Builtins{},
			Name:   "Closure",
		},
		Name: "create",
	}
	for _, test := range []struct {
		name    string
		in, out shell.Expr
	}{
		{
			name: "Var",
			in:   shell.VarRef{Name: "x"},
			out:  shell.VarRef{Name: "x"},
		},
		{
			name: "NoClosure",
			in: shell.Lambda{Args: []string{"x"}, Body: []shell.Stmt{
				shell.ExprStmt{Expr: shell.VarRef{Name: "x"}},
			}},
			out: shell.Lambda{Args: []string{"x"}, Body: []shell.Stmt{
				shell.ExprStmt{Expr: shell.VarRef{Name: "x"}},
			}},
		},
		{
			name: "Closure",
			in: shell.Lambda{Body: []shell.Stmt{
				shell.ExprStmt{Expr: shell.VarRef{Name: "x"}},
				shell.ExprStmt{Expr: shell.VarRef{Name: "y"}},
				shell.ExprStmt{Expr: shell.VarRef{Name: "z"}},
			}},
			out: shell.Call{
				Fn: closure,
				Args: []shell.Expr{
					shell.Lambda{Args: []string{"x", "y", "z"}, Body: []shell.Stmt{
						shell.ExprStmt{Expr: shell.VarRef{Name: "x"}},
						shell.ExprStmt{Expr: shell.VarRef{Name: "y"}},
						shell.ExprStmt{Expr: shell.VarRef{Name: "z"}},
					}},
					shell.VarRef{Name: "x"},
					shell.VarRef{Name: "y"},
					shell.VarRef{Name: "z"},
				},
			},
		},
		{
			name: "ArgClosure",
			in: shell.Lambda{Args: []string{"x"}, Body: []shell.Stmt{
				shell.ExprStmt{Expr: shell.VarRef{Name: "x"}},
				shell.ExprStmt{Expr: shell.VarRef{Name: "y"}},
				shell.ExprStmt{Expr: shell.VarRef{Name: "z"}},
			}},
			out: shell.Call{
				Fn: closure,
				Args: []shell.Expr{
					shell.Lambda{Args: []string{"x", "y", "z"}, Body: []shell.Stmt{
						shell.ExprStmt{Expr: shell.VarRef{Name: "x"}},
						shell.ExprStmt{Expr: shell.VarRef{Name: "y"}},
						shell.ExprStmt{Expr: shell.VarRef{Name: "z"}},
					}},
					shell.VarRef{Name: "y"},
					shell.VarRef{Name: "z"},
				},
			},
		},
		{
			name: "DefClosure",
			in: shell.Lambda{Body: []shell.Stmt{
				shell.VarDef{Name: "y", Value: shell.IntLit{Value: 2}},
				shell.ExprStmt{Expr: shell.VarRef{Name: "x"}},
				shell.ExprStmt{Expr: shell.VarRef{Name: "y"}},
				shell.ExprStmt{Expr: shell.VarRef{Name: "z"}},
			}},
			out: shell.Call{
				Fn: closure,
				Args: []shell.Expr{
					shell.Lambda{Args: []string{"x", "z"}, Body: []shell.Stmt{
						shell.VarDef{Name: "y", Value: shell.IntLit{Value: 2}},
						shell.ExprStmt{Expr: shell.VarRef{Name: "x"}},
						shell.ExprStmt{Expr: shell.VarRef{Name: "y"}},
						shell.ExprStmt{Expr: shell.VarRef{Name: "z"}},
					}},
					shell.VarRef{Name: "x"},
					shell.VarRef{Name: "z"},
				},
			},
		},
		{
			name: "RedefClosure",
			in: shell.Lambda{Body: []shell.Stmt{
				shell.VarDef{Name: "y", Value: shell.VarRef{Name: "y"}},
				shell.ExprStmt{Expr: shell.VarRef{Name: "x"}},
				shell.ExprStmt{Expr: shell.VarRef{Name: "y"}},
				shell.ExprStmt{Expr: shell.VarRef{Name: "z"}},
			}},
			out: shell.Call{
				Fn: closure,
				Args: []shell.Expr{
					shell.Lambda{Args: []string{"x", "y", "z"}, Body: []shell.Stmt{
						shell.VarDef{Name: "y", Value: shell.VarRef{Name: "y"}},
						shell.ExprStmt{Expr: shell.VarRef{Name: "x"}},
						shell.ExprStmt{Expr: shell.VarRef{Name: "y"}},
						shell.ExprStmt{Expr: shell.VarRef{Name: "z"}},
					}},
					shell.VarRef{Name: "x"},
					shell.VarRef{Name: "y"},
					shell.VarRef{Name: "z"},
				},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.out, removeClosures(test.in))
		})
	}
}
