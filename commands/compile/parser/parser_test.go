package parser

import (
	"testing"

	"github.com/bobappleyard/cezanne/commands/compile/ast"
	"github.com/bobappleyard/cezanne/format/symtab"
	"github.com/bobappleyard/cezanne/util/assert"
)

func TestParseFile(t *testing.T) {
	var syms symtab.Symtab

	for _, test := range []struct {
		name string
		in   string
		out  ast.Package
	}{
		{
			name: "IntLit",
			in:   `func main() {1}`,
			out: ast.Package{
				Name:    symtab.Symbol{},
				Imports: []ast.Import{},
				Funcs: []ast.Method{{
					Name: syms.SymbolID("main"),
					Args: []symtab.Symbol{},
					Body: ast.Int{Value: 1},
				}},
				Vars: []ast.Var{},
			},
		},
		{
			name: "StrLit",
			in:   `func main() {"hello"}`,
			out: ast.Package{
				Name:    symtab.Symbol{},
				Imports: []ast.Import{},
				Funcs: []ast.Method{{
					Name: syms.SymbolID("main"),
					Args: []symtab.Symbol{},
					Body: ast.String{Value: "hello"},
				}},
				Vars: []ast.Var{},
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
				Name:    symtab.Symbol{},
				Imports: []ast.Import{},
				Funcs: []ast.Method{{
					Name: syms.SymbolID("main"),
					Args: []symtab.Symbol{},
					Body: ast.Create{Methods: []ast.Method{{
						Name: syms.SymbolID("visit"),
						Args: []symtab.Symbol{syms.SymbolID("v")},
						Body: ast.Ref{Name: syms.SymbolID("v")},
					}}},
				}},
				Vars: []ast.Var{},
			},
		},
		{
			name: "ObjectManyMethods",
			in: `

				func main() {
					object {
						true() { v }
						false() { u }
					}
				}
				`,
			out: ast.Package{
				Name:    symtab.Symbol{},
				Imports: []ast.Import{},
				Funcs: []ast.Method{{
					Name: syms.SymbolID("main"),
					Args: []symtab.Symbol{},
					Body: ast.Create{Methods: []ast.Method{
						{
							Name: syms.SymbolID("true"),
							Args: []symtab.Symbol{},
							Body: ast.Ref{Name: syms.SymbolID("v")},
						},
						{
							Name: syms.SymbolID("false"),
							Args: []symtab.Symbol{},
							Body: ast.Ref{Name: syms.SymbolID("u")},
						},
					}},
				}},
				Vars: []ast.Var{},
			},
		},
		{
			name: "NestedObject",
			in: `
			func main() {
				test.match(object {
					true() {
						1
					}
					false() {
						2
					}
				})
			}
			`,
			out: ast.Package{
				Name:    symtab.Symbol{},
				Imports: []ast.Import{},
				Funcs: []ast.Method{{
					Name: syms.SymbolID("main"),
					Args: []symtab.Symbol{},
					Body: ast.Invoke{
						Object: ast.Ref{Name: syms.SymbolID("test")},
						Name:   syms.SymbolID("match"),
						Args: []ast.Expr{ast.Create{Methods: []ast.Method{
							{
								Name: syms.SymbolID("true"),
								Args: []symtab.Symbol{},
								Body: ast.Int{Value: 1},
							},
							{
								Name: syms.SymbolID("false"),
								Args: []symtab.Symbol{},
								Body: ast.Int{Value: 2},
							},
						}}}},
				}},
				Vars: []ast.Var{},
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
				Name:    symtab.Symbol{},
				Imports: []ast.Import{},
				Funcs: []ast.Method{{
					Name: syms.SymbolID("main"),
					Args: []symtab.Symbol{},
					Body: ast.Invoke{
						Object: ast.Ref{Name: syms.SymbolID("x")},
						Name:   syms.SymbolID("method"),
						Args:   []ast.Expr{ast.Int{Value: 1}, ast.Int{Value: 2}}},
				}},
				Vars: []ast.Var{},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var m ast.Package
			err := ParseFile(&syms, &m, []byte(test.in))
			assert.Nil(t, err)
			assert.Equal(t, m, test.out)
		})
	}

	t.Log(syms)
}

func TestFullParse(t *testing.T) {
	var syms symtab.Symtab

	var m ast.Package
	err := ParseFile(&syms, &m, []byte(`
	func main() {
		handle trigger Write(2) {
			Write(x) { resume.call(x) }
		}
	}
	`))
	assert.Nil(t, err)
	assert.Equal(t, ast.Package{
		Name:    symtab.Symbol{},
		Imports: []ast.Import{},
		Funcs: []ast.Method{{
			Name: syms.SymbolID("main"),
			Args: []symtab.Symbol{},
			Body: ast.Handle{
				In: ast.Trigger{Name: syms.SymbolID("Write"), Args: []ast.Expr{ast.Int{Value: 2}}},
				With: []ast.Method{{
					Name: syms.SymbolID("Write"),
					Args: []symtab.Symbol{syms.SymbolID("context"), syms.SymbolID("x")},
					Body: ast.Invoke{
						Object: ast.Ref{Name: syms.SymbolID("resume")},
						Name:   syms.SymbolID("call"),
						Args:   []ast.Expr{ast.Ref{Name: syms.SymbolID("x")}},
					}},
				},
			},
		}},
		Vars: []ast.Var{},
	}, m)

}
