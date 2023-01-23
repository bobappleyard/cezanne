package assembly

import (
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/format"
)

func TestSimpleInstructions(t *testing.T) {
	var b Block
	b.Load(1)
	b.Store(2)
	b.Natural(b.Fixed(3))
	b.Return()

	p := b.Package()
	assert.Equal(t, []byte{
		format.LoadOp, 1,
		format.StoreOp, 2,
		format.NaturalOp, 3, 0, 0, 0,
		format.RetOp,
	}, p.Code)
}

func TestLocation(t *testing.T) {
	var b Block

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
	assert.Equal(t, []byte{
		format.NaturalOp, 0, 0, 0, 0,
		format.StoreOp, 1,
		format.NaturalOp, 0, 0, 0, 0,
		format.StoreOp, 2,
		format.NaturalOp, 3, 0, 0, 0,
	}, p.Code)
	assert.Equal(t, p.Relocations, []format.Relocation{
		{Kind: format.CodeRel, ID: 14, Pos: 1},
		{Kind: format.CodeRel, ID: 19, Pos: 8},
	})
}

func TestImport(t *testing.T) {
	var b Block

	b.GlobalLoad(b.Global("fmt"))

	p := b.Package()
	assert.Equal(t, []byte{
		format.GlobalLoadOp, 0, 0, 0, 0,
	}, p.Code)
	assert.Equal(t, []string{"fmt"}, p.Imports)
	assert.Equal(t, []format.Relocation{
		{Kind: format.GlobalRel, ID: 0, Pos: 1},
	}, p.Relocations)
}

func TestClass(t *testing.T) {
	var b Block

	b.Create(b.Class("Point"), 2)

	p := b.Package()
	assert.Equal(t, []byte{
		format.CreateOp, 0, 0, 0, 0, 2,
	}, p.Code)
	assert.Equal(t, []format.Class{{
		Name: "Point",
	}}, p.Classes)
	assert.Equal(t, []format.Relocation{
		{Kind: format.ClassRel, ID: 0, Pos: 1},
	}, p.Relocations)
}

func TestMethod(t *testing.T) {
	var b Block

	b.Call(b.Method("add"), 4)

	p := b.Package()
	assert.Equal(t, []byte{
		format.CallOp, 0, 0, 0, 0, 4,
	}, p.Code)
	assert.Equal(t, []format.Method{{
		Name:       "add",
		Visibility: format.Public,
	}}, p.Methods)
	assert.Equal(t, []format.Relocation{
		{Kind: format.MethodRel, ID: 0, Pos: 1},
	}, p.Relocations)
}

func TestBinding(t *testing.T) {
	var b Block

	b.Create(b.Class("MainPackage"), 0)
	b.Call(b.Method("main"), 0)
	b.ImplementMethod(b.Class("MainPackage"), b.Method("main"))

	p := b.Package()
	assert.Equal(t, []format.Implementation{{
		Kind:       format.StandardBinding,
		EntryPoint: 12,
	}}, p.Implementations)

}
