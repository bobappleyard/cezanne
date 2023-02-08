package backend

import (
	"fmt"
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/format/assembly"
	"github.com/bobappleyard/cezanne/linker"
	"github.com/bobappleyard/cezanne/runtime/api"
	"github.com/bobappleyard/cezanne/runtime/env"
)

type testLinkerEnv map[string]*format.Package

// LoadPackage implements linker.LinkerEnv
func (e testLinkerEnv) LoadPackage(path string) (*format.Package, error) {
	pkg := e[path]
	if pkg == nil {
		return nil, fmt.Errorf("%s: %w", path, linker.ErrMissingPackage)
	}
	return pkg, nil
}

func TestBuildPackage(t *testing.T) {

	pkg, err := BuildPackage(ast.Package{
		Name: "main",
		Imports: []ast.Import{
			{Name: "test", Path: "test"},
			{Name: "#pkg", Path: "."},
		},
		Funcs: []ast.Method{
			{
				Name: "main",
				Body: ast.Invoke{
					Object: ast.Ref{Name: "test"},
					Name:   "print",
					Args: []ast.Expr{
						ast.Invoke{
							Object: ast.Ref{Name: "#pkg"},
							Name:   "fac",
							Args:   []ast.Expr{ast.Int{Value: 4}},
						},
					},
				},
			},
			{
				Name: "fac",
				Args: []string{"x"},
				Body: ast.Invoke{
					Object: ast.Invoke{
						Object: ast.Ref{Name: "test"},
						Name:   "lte",
						Args: []ast.Expr{
							ast.Ref{Name: "x"},
							ast.Int{Value: 1},
						},
					},
					Name: "match",
					Args: []ast.Expr{ast.Create{Methods: []ast.Method{
						{
							Name: "true",
							Body: ast.Ref{Name: "x"},
						},
						{
							Name: "false",
							Body: ast.Invoke{
								Object: ast.Ref{Name: "test"},
								Name:   "mul",
								Args: []ast.Expr{
									ast.Ref{Name: "x"},
									ast.Invoke{
										Object: ast.Ref{Name: "#pkg"},
										Name:   "fac",
										Args: []ast.Expr{
											ast.Invoke{
												Object: ast.Ref{Name: "test"},
												Name:   "sub",
												Args: []ast.Expr{
													ast.Ref{Name: "x"},
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

	prog, err := linker.Link(testLinkerEnv{
		"main": pkg,
		"test": testPkg(),
	})
	assert.Nil(t, err)

	t.Log(pkg)
	t.Log(prog)
	e := new(env.Env)
	e.SetHeapSize(32)

	var logged []api.Object
	e.AddExternalMethod("test:print", func(p *env.Thread, recv api.Object) {
		logged = append(logged, p.Arg(0))
		p.Return(env.Int(0))
	})

	var trueM, falseM format.MethodID
	for i, m := range prog.Methods {
		if m.Name == "true" {
			trueM = format.MethodID(i)
		}
		if m.Name == "false" {
			falseM = format.MethodID(i)
		}
	}

	e.AddExternalMethod("test:lte", func(p *env.Thread, recv api.Object) {
		if env.AsInt(p.Arg(0)) <= env.AsInt(p.Arg(1)) {
			p.TailCall(recv, trueM)
		} else {
			p.TailCall(recv, falseM)
		}
	})

	e.AddExternalMethod("test:sub", func(p *env.Thread, recv api.Object) {
		p.Return(env.Int(env.AsInt(p.Arg(0)) - env.AsInt(p.Arg(1))))
	})

	e.AddExternalMethod("test:mul", func(p *env.Thread, recv api.Object) {
		p.Return(env.Int(env.AsInt(p.Arg(0)) * env.AsInt(p.Arg(1))))
	})

	e.Run(prog)
	t.Log(logged)

	assert.Equal(t, logged, []api.Object{env.Int(24)})

}

func testPkg() *format.Package {
	var b assembly.Package

	pkgClass := b.Class(0)

	b.Create(pkgClass, 0)
	b.Return()

	trueClass := b.Class(0)
	b.ImplementMethod(trueClass, b.Method("match"))
	b.Load(2)
	b.Call(b.Method("true"), 0)

	b.ImplementMethod(pkgClass, b.Method("true"))
	b.Create(trueClass, 0)
	b.Return()

	falseClass := b.Class(0)
	b.ImplementMethod(falseClass, b.Method("match"))
	b.Load(2)
	b.Call(b.Method("false"), 0)

	b.ImplementMethod(pkgClass, b.Method("false"))
	b.Create(falseClass, 0)
	b.Return()

	b.ImplementExternalMethod(pkgClass, b.Method("lte"), "test:lte")
	b.ImplementExternalMethod(pkgClass, b.Method("sub"), "test:sub")
	b.ImplementExternalMethod(pkgClass, b.Method("mul"), "test:mul")
	b.ImplementExternalMethod(pkgClass, b.Method("print"), "test:print")

	return b.Package()
}
