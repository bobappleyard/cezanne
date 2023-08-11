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
			in:   `func main() -> 1`,
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

				func main() -> {
					visit(v) -> v
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
			name: "ObjectManyMethods",
			in: `

				func main() -> {
					true() -> v 
					false() -> u
				}
				`,
			out: ast.Package{
				Funcs: []ast.Method{{
					Name: "main",
					Body: ast.Create{
						Methods: []ast.Method{
							{
								Name: "true",
								Body: ast.Ref{Name: "v"},
							},
							{
								Name: "false",
								Body: ast.Ref{Name: "u"},
							},
						},
					},
				}},
			},
		},
		{
			name: "NestedObject",
			in: `
			func main() -> test.match({
				true() -> 1
				false() -> 2
			})
			`,
			out: ast.Package{
				Funcs: []ast.Method{{Name: "main", Body: ast.Invoke{
					Object: ast.Ref{Name: "test"},
					Name:   "match",
					Args: []ast.Expr{
						ast.Create{Methods: []ast.Method{
							{
								Name: "true",
								Body: ast.Int{Value: 1},
							},
							{
								Name: "false",
								Body: ast.Int{Value: 2},
							},
						}},
					},
				}}},
			},
		},
		{
			name: "MultilineParams",
			in: `

				func main() -> x.method(
					1,
					2
				)

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
		{
			name: "Currying",
			in: `

				func fold(f) -> (i) -> (l) -> jump(f, i, l)
				
			`,
			out: ast.Package{
				Funcs: []ast.Method{{
					Name: "fold",
					Args: []string{"f"},
					Body: ast.Create{
						Methods: []ast.Method{{
							Name: "call",
							Args: []string{"i"},
							Body: ast.Create{
								Methods: []ast.Method{{
									Name: "call",
									Args: []string{"l"},
									Body: ast.Invoke{
										Object: ast.Ref{Name: "jump"},
										Name:   "call",
										Args: []ast.Expr{
											ast.Ref{Name: "f"},
											ast.Ref{Name: "i"},
											ast.Ref{Name: "l"},
										},
									},
								}},
							},
						}},
					},
				}},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var m ast.Package
			err := ParseFile(&m, []byte(test.in))
			assert.Nil(t, err)
			assert.Equal(t, m, test.out)
		})
	}
}

func TestFullParse(t *testing.T) {
	var m ast.Package
	err := ParseFile(&m, []byte(`
	func main() -> handle trigger Write(2) {
		Write(x) -> resume.call(x)
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
