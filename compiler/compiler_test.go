package compiler

import (
	"fmt"
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/compiler/backend"
	"github.com/bobappleyard/cezanne/compiler/parser"
	"github.com/bobappleyard/cezanne/cz-link/linker"
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/format/assembly"
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

func TestCompiler(t *testing.T) {
	var pt ast.Package

	err := parser.ParseFile(&pt, []byte(`

	import test

	func main() {
		test.print(fac(5))
	}

	func fac(n) { 
		test.lte(n, 1).match(object {
			true() { 1 }
			false() { test.mul(n, fac(test.sub(n, 1))) }
		})
	}

	`))
	assert.Nil(t, err)

	pkg, err := backend.BuildPackage(pt)
	assert.Nil(t, err)

	t.Log(pkg)

	e := new(env.Env)
	e.SetHeapSize(32)

	prog, err := linker.Link(testLinkerEnv{
		"main": pkg,
		"test": testPkg(),
	})
	assert.Nil(t, err)
	t.Log(prog)

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

	e.Run(prog)

	assert.Equal(t, logged, []api.Object{{Class: prog.CoreKinds[format.IntKind], Data: 120}})
}

func testPkg() *format.Package {
	var b assembly.Writer

	pkgClass := b.Class(0)

	b.Create(pkgClass, 0)
	b.Return()

	trueClass := b.Class(0)
	b.ImplementMethod(trueClass, b.Method("match"))
	b.Load(2)
	b.Call(b.Method("true"), 0)

	falseClass := b.Class(0)
	b.ImplementMethod(falseClass, b.Method("match"))
	b.Load(2)
	b.Call(b.Method("false"), 0)

	b.ImplementExternalMethod(pkgClass, b.Method("lte"), "test:lte")
	b.ImplementExternalMethod(pkgClass, b.Method("sub"), "test:sub")
	b.ImplementExternalMethod(pkgClass, b.Method("mul"), "test:mul")
	b.ImplementExternalMethod(pkgClass, b.Method("print"), "test:print")

	p := b.Package()
	p.Classes[1].Kind = format.TrueKind
	p.Classes[2].Kind = format.FalseKind

	return p

}