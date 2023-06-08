package link

import (
	"testing"

	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/format/assembly"
	"github.com/bobappleyard/cezanne/format/symtab"
	"github.com/bobappleyard/cezanne/runtime/api"
	"github.com/bobappleyard/cezanne/runtime/env"
	"github.com/bobappleyard/cezanne/util/assert"
)

type mockLinkerEnv map[string]*format.Package

func (e mockLinkerEnv) LoadPackage(path string) (*format.Package, error) {
	if p, ok := e[path]; ok {
		return p, nil
	}
	return nil, ErrMissingPackage
}

func TestCircularImports(t *testing.T) {
	var syms symtab.Symtab

	prog, err := Link(&syms, mockLinkerEnv{
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
	var syms symtab.Symtab

	prog, err := Link(&syms, mockLinkerEnv{
		"main": mainPackage(&syms),
		"dep":  depPackage(&syms),
		"core": corePackage(&syms),
	})
	assert.Nil(t, err)
	t.Log(prog)

	e := new(env.Env)
	e.SetHeapSize(32)

	var res api.Object
	e.AddExternalMethod("test:result", func(p *env.Thread, recv api.Object) {
		res = p.Arg(0)
		p.Return(p.Process().Int(0))
	})

	e.AddExternalMethod("core:int_add", func(p *env.Thread, recv api.Object) {
		p.Return(p.Process().Int(p.Process().AsInt(p.Arg(0)) + p.Process().AsInt(p.Arg(1))))
	})

	e.Run(&syms, prog)

	assert.Equal(t, res, api.Object{Class: prog.CoreKinds[format.IntKind], Data: 10})
}

func mainPackage(syms *symtab.Symtab) *format.Package {
	var b assembly.Writer

	k := b.Location()
	core := b.Global(0)
	pkg := b.Class(0)

	b.Create(pkg, 0)
	b.Return()

	b.ImplementMethod(pkg, b.Method(syms.SymbolID("main")))
	b.GlobalLoad(b.Import("core"))
	b.GlobalStore(core)
	b.Natural(b.Fixed(2))
	b.Store(2)
	b.Natural(k)
	b.Store(3)
	b.Natural(b.Fixed(5))
	b.Store(4)
	b.GlobalLoad(b.Import("dep"))
	b.Call(b.Method(syms.SymbolID("add5")), 2)
	k.Define()
	b.Store(2)
	b.GlobalLoad(core)
	b.Call(b.Method(syms.SymbolID("result")), 0)

	return b.Package()
}

func depPackage(syms *symtab.Symtab) *format.Package {
	var b assembly.Writer

	pkg := b.Class(0)

	b.Create(pkg, 0)
	b.Return()

	b.ImplementMethod(pkg, b.Method(syms.SymbolID("add5")))
	b.Natural(b.Fixed(5))
	b.Store(3)
	b.GlobalLoad(b.Import("core"))
	b.Call(b.Method(syms.SymbolID("int_add")), 0)

	return b.Package()
}

func corePackage(syms *symtab.Symtab) *format.Package {
	var b assembly.Writer

	pkg := b.Class(0)

	b.Create(pkg, 0)
	b.Return()

	b.ImplementExternalMethod(pkg, b.Method(syms.SymbolID("result")), syms.SymbolID("test:result"))
	b.ImplementExternalMethod(pkg, b.Method(syms.SymbolID("int_add")), syms.SymbolID("core:int_add"))

	return b.Package()
}
