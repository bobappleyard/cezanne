package backend

import (
	"testing"

	"github.com/bobappleyard/cezanne/format/assembly"
	"github.com/bobappleyard/cezanne/format/symtab"
	"github.com/bobappleyard/cezanne/util/assert"
)

func TestTailCall(t *testing.T) {
	var syms symtab.Symtab

	b := method{
		varc: 3,
		steps: []step{
			stringStep{val: "abc", into: 0},
			intStep{val: 1, into: 1},
			callStep{object: 0, method: syms.SymbolID("add"), into: 2, params: []variable{1}},
			returnStep{val: 2},
		},
	}

	w := &assembler{
		syms: &syms,
	}

	w.writePackage(b)

	expect := new(assembly.Writer)
	start := expect.Location()
	end := expect.Location()
	k := expect.Location()

	expect.Natural(expect.Fixed(5))
	expect.Store(5)
	expect.Natural(k)
	expect.Store(6)
	expect.Natural(start)
	expect.Store(7)
	expect.Natural(end)
	expect.Store(8)
	expect.GlobalLoad(expect.Import("runtime"))
	expect.Call(expect.Method(syms.SymbolID("string_constant")), 5)
	k.Define()
	expect.Store(2)
	expect.Natural(expect.Fixed(1))
	expect.Store(3)
	expect.Load(2)
	expect.Store(5)
	expect.Load(3)
	expect.Store(6)
	expect.Load(6)
	expect.Store(2)
	expect.Load(5)
	expect.Call(expect.Method(syms.SymbolID("add")), 0)
	start.Define()
	expect.WriteByte('a')
	expect.WriteByte('b')
	expect.WriteByte('c')
	end.Define()

	assert.Equal(t, &w.dest, expect)
}
