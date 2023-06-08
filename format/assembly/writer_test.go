package assembly

import (
	"testing"

	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/format/symtab"
	"github.com/bobappleyard/cezanne/util/assert"
)

func TestSimpleInstructions(t *testing.T) {
	var tab symtab.Symtab
	b := New(&tab)
	b.Load(1)
	b.Store(2)
	b.Natural(b.Fixed(3))
	b.Return()

	p := b.Package()
	assert.Equal(t, p.Code, []byte{
		format.LoadOp, 1,
		format.StoreOp, 2,
		format.NaturalOp, 3, 0, 0, 0,
		format.RetOp,
	})
}

func TestLocation(t *testing.T) {
	var tab symtab.Symtab
	b := New(&tab)

	start := b.Location()
	end := b.Location()

	b.Natural(start)
	b.Store(1)
	b.Natural(end)
	b.Store(2)
	start.Define()
	b.Natural(b.Fixed(3))
	end.Define()

	p := b.Package()
	assert.Equal(t, p.Code, []byte{
		format.NaturalOp, 0, 0, 0, 0,
		format.StoreOp, 1,
		format.NaturalOp, 0, 0, 0, 0,
		format.StoreOp, 2,
		format.NaturalOp, 3, 0, 0, 0,
	})
	assert.Equal(t, p.Relocations, []format.Relocation{
		{Kind: format.CodeRel, ID: 14, Pos: 1},
		{Kind: format.CodeRel, ID: 19, Pos: 8},
	})
}

func TestGlobal(t *testing.T) {
	var tab symtab.Symtab
	b := New(&tab)

	b.GlobalLoad(b.Global(0))

	p := b.Package()
	assert.Equal(t, p.Code, []byte{
		format.GlobalLoadOp, 0, 0, 0, 0,
	})
	assert.Equal(t, p.Relocations, []format.Relocation{
		{Kind: format.GlobalRel, ID: 0, Pos: 1},
	})
}

func TestImport(t *testing.T) {
	var tab symtab.Symtab
	b := New(&tab)

	b.GlobalLoad(b.Import("fmt"))

	p := b.Package()
	assert.Equal(t, p.Code, []byte{
		format.GlobalLoadOp, 0, 0, 0, 0,
	})
	assert.Equal(t, p.Imports, []string{"fmt"})
	assert.Equal(t, p.Relocations, []format.Relocation{
		{Kind: format.ImportRel, ID: 0, Pos: 1},
	})
}

func TestClass(t *testing.T) {
	var tab symtab.Symtab
	b := New(&tab)

	b.Create(b.Class(0), 2)

	p := b.Package()
	assert.Equal(t, p.Code, []byte{
		format.CreateOp, 0, 0, 0, 0, 2,
	})
	assert.Equal(t, p.Classes, []format.Class{{}})
	assert.Equal(t, p.Relocations, []format.Relocation{
		{Kind: format.ClassRel, ID: 0, Pos: 1},
	})
}

func TestMethod(t *testing.T) {
	var tab symtab.Symtab
	b := New(&tab)

	b.Call(b.Method(tab.SymbolID("add")), 4)

	p := b.Package()
	assert.Equal(t, p.Code, []byte{
		format.CallOp, 0, 0, 0, 0, 4,
	})
	assert.Equal(t, p.Methods, []format.Method{{
		Name:       tab.SymbolID("add"),
		Visibility: format.Public,
	}})
	assert.Equal(t, p.Relocations, []format.Relocation{
		{Kind: format.MethodRel, ID: 0, Pos: 1},
	})
}

func TestBinding(t *testing.T) {
	var tab symtab.Symtab
	b := New(&tab)

	mainPackage := b.Class(0)

	b.Create(mainPackage, 0)
	b.Call(b.Method(tab.SymbolID("main")), 0)
	b.ImplementMethod(mainPackage, b.Method(tab.SymbolID("main")))

	p := b.Package()
	assert.Equal(t, p.Implementations, []format.Implementation{{
		Kind:       format.StandardBinding,
		EntryPoint: 12,
	}})

}
