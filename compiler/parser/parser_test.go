package parser

import (
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/compiler/ast"
)

func TestParseFile(t *testing.T) {
	for _, test := range []struct {
		name string
		in   string
		out  ast.Package
	}{
		{
			name: "IntLit",
			in:   `func main() {1}`,
			out: ast.Package{
				Funcs: []ast.Method{{
					Name: "main",
					Body: ast.Int{Value: 1},
				}},
			},
		},
		{
			name: "Object",
			in: `

				func main() {
					object {
						visit(v) { v }
					}
				}
				`,
			out: ast.Package{
				Funcs: []ast.Method{{
					Name: "main",
					Body: ast.Create{
						Methods: []ast.Method{{
							Name: "visit",
							Args: []string{"v"},
							Body: ast.Ref{Name: "v"},
						}},
					},
				}},
			},
		},
		{
			name: "MultilineParams",
			in: `
			
				func main() {
					x.method(
						1,
						2
					)
				}
			`,
			out: ast.Package{
				Funcs: []ast.Method{{
					Name: "main",
					Body: ast.Invoke{
						Object: ast.Ref{Name: "x"},
						Name:   "method",
						Args: []ast.Expr{
							ast.Int{Value: 1},
							ast.Int{Value: 2},
						},
					},
				}},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var m ast.Package
			err := ParseFile(&m, []byte(test.in))
			assert.Nil(t, err)
			assert.Equal(t, test.out, m)
		})
	}
}

func TestFullParse(t *testing.T) {
	var m ast.Package
	err := ParseFile(&m, []byte(`
	func main() {
		handle trigger Write(2) {
			Write(x) { resume.call(x) }
		}
	}
	`))
	assert.Nil(t, err)
	assert.Equal(t, ast.Package{
		Funcs: []ast.Method{{
			Name: "main",
			Body: ast.Handle{
				In: ast.Trigger{
					Name: "Write",
					Args: []ast.Expr{ast.Int{Value: 2}},
				},
				With: []ast.Method{{
					Name: "Write",
					Args: []string{"context", "x"},
					Body: ast.Invoke{
						Object: ast.Ref{Name: "resume"},
						Name:   "call",
						Args:   []ast.Expr{ast.Ref{Name: "x"}},
					},
				}},
			},
		}},
	}, m)

}
