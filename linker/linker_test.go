package linker

import (
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/format/assembly"
	"github.com/bobappleyard/cezanne/runtime/api"
	"github.com/bobappleyard/cezanne/runtime/env"
)

type mockLinkerEnv map[string]*format.Package

func (e mockLinkerEnv) LoadPackage(path string) (*format.Package, error) {
	if p, ok := e[path]; ok {
		return p, nil
	}
	return nil, ErrMissingPackage
}

func TestCircularImports(t *testing.T) {
	prog, err := Link(mockLinkerEnv{
		"main": &format.Package{
			Imports: []string{"a"},
		},
		"a": &format.Package{
			Imports: []string{"b"},
		},
		"b": &format.Package{
			Imports: []string{"a"},
		},
	})
	assert.Nil(t, prog)
	assert.Equal(t, err, ErrCircularImport)
}

func TestLink(t *testing.T) {
	prog, err := Link(mockLinkerEnv{
		"main": mainPackage(),
		"dep":  depPackage(),
		"core": corePackage(),
	})
	assert.Nil(t, err)
	t.Log(prog)

	e := new(env.Env)
	e.SetHeapSize(32)

	var res api.Object
	e.AddExternalMethod("test:result", func(p *env.Thread, recv api.Object) {
		res = p.Arg(0)
		p.Return(env.Int(0))
	})

	e.AddExternalMethod("core:int_add", func(p *env.Thread, recv api.Object) {
		p.Return(env.Int(env.AsInt(p.Arg(0)) + env.AsInt(p.Arg(1))))
	})

	e.Run(prog)

	assert.Equal(t, res, env.Int(10))
}

func mainPackage() *format.Package {
	var b assembly.Package

	k := b.Location()
	core := b.Global()

	b.GlobalLoad(b.Import("core"))
	b.GlobalStore(core)
	b.Natural(b.Fixed(2))
	b.Store(2)
	b.Natural(k)
	b.Store(3)
	b.Natural(b.Fixed(5))
	b.Store(4)
	b.GlobalLoad(b.Import("dep"))
	b.Call(b.Method("add5"), 2)
	k.Define()
	b.Store(2)
	b.GlobalLoad(core)
	b.Call(b.Method("result"), 0)

	return b.Package()
}

func depPackage() *format.Package {
	var b assembly.Package

	pkg := b.Class(0)

	b.Create(pkg, 0)
	b.Return()

	b.ImplementMethod(pkg, b.Method("add5"))
	b.Natural(b.Fixed(5))
	b.Store(3)
	b.GlobalLoad(b.Import("core"))
	b.Call(b.Method("int_add"), 0)

	return b.Package()
}

func corePackage() *format.Package {
	var b assembly.Package

	pkg := b.Class(0)

	b.Create(pkg, 0)
	b.Return()

	b.ImplementExternalMethod(pkg, b.Method("result"), "test:result")
	b.ImplementExternalMethod(pkg, b.Method("int_add"), "core:int_add")

	return b.Package()
}
