package backend

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/sexpr"
)

func TestRun(t *testing.T) {

	add := func(x, y ast.Expr) ast.Expr {
		return ast.Invoke{
			Object: ast.Ref{Name: "arith"},
			Name:   "add",
			Args:   []ast.Expr{x, y},
		}
	}

	val := func(x int) ast.Expr {
		return ast.Int{Value: x}
	}

	for _, test := range []struct {
		name   string
		prog   ast.Expr
		output string
	}{
		{
			name:   "SimpleAdd",
			prog:   add(val(2), val(2)),
			output: "4",
		},
		{
			name: "HandleEscape",
			prog: ast.Handle{
				In: add(val(2), ast.Trigger{Name: "err", Args: []ast.Expr{val(2)}}),
				With: []ast.Method{{
					Name: "err",
					Args: []string{"resume", "e"},
					Body: add(ast.Ref{Name: "e"}, val(1)),
				}},
			},
			output: "3",
		},
		{
			name: "HandleResume",
			prog: ast.Handle{
				In: add(val(2), ast.Trigger{Name: "err", Args: []ast.Expr{val(2)}}),
				With: []ast.Method{{
					Name: "err",
					Args: []string{"context", "e"},
					Body: ast.Invoke{
						Object: ast.Ref{Name: "context"},
						Name:   "resume",
						Args:   []ast.Expr{add(ast.Ref{Name: "e"}, val(3))},
					},
				}},
			},
			output: "7",
		},
		{
			name: "HandleDoubleResume",
			prog: ast.Handle{
				In: add(val(2), ast.Trigger{Name: "err", Args: []ast.Expr{val(2)}}),
				With: []ast.Method{{
					Name: "err",
					Args: []string{"context", "e"},
					Body: add(ast.Invoke{
						Object: ast.Ref{Name: "context"},
						Name:   "resume",
						Args:   []ast.Expr{add(ast.Ref{Name: "e"}, val(3))},
					}, ast.Invoke{
						Object: ast.Ref{Name: "context"},
						Name:   "resume",
						Args:   []ast.Expr{add(ast.Ref{Name: "e"}, val(2))},
					}),
				}},
			},
			output: "13",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			prog, err := os.CreateTemp("", "")
			assert.Nil(t, err)

			ctx := Context{
				module: new(Module),
			}

			s := ctx.CompileExpr(test.prog)
			ctx.defineRuntimeTypes()

			fmt.Fprintln(prog, `(load "runtime.ss")`)
			assert.Nil(t, err)
			for _, x := range ctx.Render() {
				fmt.Fprintln(prog, x)
			}

			ep := sexpr.Call("run", sexpr.Call("lambda", sexpr.List(), s))
			_, err = fmt.Fprintln(prog, sexpr.Call("write", ep))
			assert.Nil(t, err)

			prog.Close()

			t.Log(prog.Name())

			cmd := exec.Command("/usr/bin/chezscheme", "--script", prog.Name())

			out, err := cmd.CombinedOutput()
			assert.Nil(t, err)
			assert.Equal(t, test.output, string(out))
		})
	}

}
