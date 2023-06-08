package backend

import (
	"fmt"
	"testing"

	"github.com/bobappleyard/cezanne/commands/compile/ast"
	"github.com/bobappleyard/cezanne/commands/link"
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/format/assembly"
	"github.com/bobappleyard/cezanne/format/symtab"
	"github.com/bobappleyard/cezanne/runtime/api"
	"github.com/bobappleyard/cezanne/runtime/env"
	"github.com/bobappleyard/cezanne/util/assert"
)

type testLinkerEnv map[string]*format.Package

// LoadPackage implements linker.LinkerEnv
func (e testLinkerEnv) LoadPackage(path string) (*format.Package, error) {
	pkg := e[path]
	if pkg == nil {
		return nil, fmt.Errorf("%s: %w", path, link.ErrMissingPackage)
	}
	return pkg, nil
}

func TestBuildPackage(t *testing.T) {

	var syms symtab.Symtab

	pkg, err := BuildPackage(&syms, ast.Package{
		Name: syms.SymbolID("main"),
		Imports: []ast.Import{
			{Name: syms.SymbolID("test"), Path: "test"},
		},
		Funcs: []ast.Method{
			{
				Name: syms.SymbolID("main"),
				Body: ast.Invoke{
					Object: ast.Ref{Name: syms.SymbolID("test")},
					Name:   syms.SymbolID("print"),
					Args: []ast.Expr{
						ast.Invoke{
							Object: ast.Ref{Name: syms.SymbolID("fac")},
							Name:   syms.SymbolID("call"),
							Args:   []ast.Expr{ast.Int{Value: 4}},
						},
					},
				},
			},
			{
				Name: syms.SymbolID("fac"),
				Args: []symtab.Symbol{syms.SymbolID("x")},
				Body: ast.Invoke{
					Object: ast.Invoke{
						Object: ast.Ref{Name: syms.SymbolID("test")},
						Name:   syms.SymbolID("lte"),
						Args: []ast.Expr{
							ast.Ref{Name: syms.SymbolID("x")},
							ast.Int{Value: 1},
						},
					},
					Name: syms.SymbolID("match"),
					Args: []ast.Expr{ast.Create{Methods: []ast.Method{
						{
							Name: syms.SymbolID("true"),
							Body: ast.Ref{Name: syms.SymbolID("x")},
						},
						{
							Name: syms.SymbolID("false"),
							Body: ast.Invoke{
								Object: ast.Ref{Name: syms.SymbolID("test")},
								Name:   syms.SymbolID("mul"),
								Args: []ast.Expr{
									ast.Ref{Name: syms.SymbolID("x")},
									ast.Invoke{
										Object: ast.Ref{Name: syms.SymbolID("fac")},
										Name:   syms.SymbolID("call"),
										Args: []ast.Expr{
											ast.Invoke{
												Object: ast.Ref{Name: syms.SymbolID("test")},
												Name:   syms.SymbolID("sub"),
												Args: []ast.Expr{
													ast.Ref{Name: syms.SymbolID("x")},
													ast.Int{Value: 1},
												},
											},
										},
									},
								},
							},
						},
					}}},
				},
			},
		},
	})
	assert.Nil(t, err)

	prog, err := link.Link(&syms, testLinkerEnv{
		"main": pkg,
		"test": testPkg(&syms),
	})
	assert.Nil(t, err)

	t.Log(pkg)
	t.Log(prog)
	e := new(env.Env)
	e.SetHeapSize(32)

	var logged []api.Object
	e.AddExternalMethod("test:print", func(p *env.Thread, recv api.Object) {
		logged = append(logged, p.Arg(0))
		p.Return(p.Process().Int(0))
	})

	e.AddExternalMethod("test:lte", func(p *env.Thread, recv api.Object) {
		a, b := p.Process().AsInt(p.Arg(0)), p.Process().AsInt(p.Arg(1))
		p.Return(p.Process().Bool(a <= b))
	})

	e.AddExternalMethod("test:sub", func(p *env.Thread, recv api.Object) {
		a, b := p.Process().AsInt(p.Arg(0)), p.Process().AsInt(p.Arg(1))
		p.Return(p.Process().Int(a - b))
	})

	e.AddExternalMethod("test:mul", func(p *env.Thread, recv api.Object) {
		a, b := p.Process().AsInt(p.Arg(0)), p.Process().AsInt(p.Arg(1))
		p.Return(p.Process().Int(a * b))
	})

	e.Run(&syms, prog)
	t.Log(logged)

	assert.Equal(t, logged, []api.Object{{Class: prog.CoreKinds[format.IntKind], Data: 24}})

}

func testPkg(syms *symtab.Symtab) *format.Package {
	var b assembly.Writer

	pkgClass := b.Class(0)

	b.Create(pkgClass, 0)
	b.Return()

	trueClass := b.Class(0)
	b.ImplementMethod(trueClass, b.Method(syms.SymbolID("match")))
	b.Load(2)
	b.Call(b.Method(syms.SymbolID("true")), 0)

	b.ImplementMethod(pkgClass, b.Method(syms.SymbolID("true")))
	b.Create(trueClass, 0)
	b.Return()

	falseClass := b.Class(0)
	b.ImplementMethod(falseClass, b.Method(syms.SymbolID("match")))
	b.Load(2)
	b.Call(b.Method(syms.SymbolID("false")), 0)

	b.ImplementMethod(pkgClass, b.Method(syms.SymbolID("false")))
	b.Create(falseClass, 0)
	b.Return()

	b.ImplementExternalMethod(pkgClass, b.Method(syms.SymbolID("lte")), syms.SymbolID("test:lte"))
	b.ImplementExternalMethod(pkgClass, b.Method(syms.SymbolID("sub")), syms.SymbolID("test:sub"))
	b.ImplementExternalMethod(pkgClass, b.Method(syms.SymbolID("mul")), syms.SymbolID("test:mul"))
	b.ImplementExternalMethod(pkgClass, b.Method(syms.SymbolID("print")), syms.SymbolID("test:print"))

	p := b.Package()
	p.Classes[1].Kind = format.TrueKind
	p.Classes[2].Kind = format.FalseKind

	return p
}
