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

	var res api.Object
	ext := map[string]func(*env.Process){
		"test:result": func(p *env.Process) {
			res = p.Arg(0)
			p.Return(env.Int(0))
		},
		"core:int_add": func(p *env.Process) {
			p.Return(env.Int(env.AsInt(p.Arg(0)) + env.AsInt(p.Arg(1))))
		},
	}
	e := env.New(32)
	e.Load(ext, prog)
	e.Run()

	assert.Equal(t, res, env.Int(10))
}

func mainPackage() *format.Package {
	var b assembly.Block

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
	var b assembly.Block

	b.Create(b.Class("Package"), 0)
	b.Return()

	b.ImplementMethod(b.Class("Package"), b.Method("add5"))
	b.Natural(b.Fixed(5))
	b.Store(3)
	b.GlobalLoad(b.Import("core"))
	b.Call(b.Method("int_add"), 0)

	return b.Package()
}

func corePackage() *format.Package {
	var b assembly.Block

	b.Create(b.Class("Package"), 0)
	b.Return()

	b.ImplementExternalMethod(b.Class("Package"), b.Method("result"), "test:result")
	b.ImplementExternalMethod(b.Class("Package"), b.Method("int_add"), "core:int_add")

	return b.Package()
}
