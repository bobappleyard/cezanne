package env

import (
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/runtime/api"
	"github.com/bobappleyard/cezanne/runtime/memory"
)

func newTestProc() *Process {
	p := &Process{}
	p.memory = memory.NewArena(p, 32)
	return p
}

func TestLoad(t *testing.T) {
	e := newTestProc()
	e.code = []byte{
		format.LoadOp, 0,
	}
	p := &Thread{process: e}
	p.data[0] = Int(25)

	p.step()

	assert.Equal(t, p.value, Int(25))
}

func TestStore(t *testing.T) {
	e := newTestProc()
	e.code = []byte{
		format.StoreOp, 0,
	}
	p := &Thread{process: e}
	p.value = Int(25)

	p.step()

	assert.Equal(t, p.data[0], Int(25))
}

func TestNatural(t *testing.T) {
	e := newTestProc()
	e.code = []byte{
		format.NaturalOp, 25, 0, 0, 0,
	}
	p := &Thread{process: e}

	p.step()

	assert.Equal(t, Int(25), p.value)
}

func TestLoadGlobal(t *testing.T) {
	e := newTestProc()
	e.code = []byte{
		format.GlobalLoadOp, 0, 0, 0, 0,
	}
	e.globals = []api.Object{
		Int(5),
	}

	p := &Thread{process: e}

	p.step()

	assert.Equal(t, p.value, Int(5))
}

func TestStoreGlobal(t *testing.T) {
	e := newTestProc()
	e.code = []byte{
		format.GlobalStoreOp, 0, 0, 0, 0,
	}
	e.globals = make([]api.Object, 1)

	p := &Thread{process: e}
	p.value = Int(25)
	p.step()

	assert.Equal(t, e.globals, []api.Object{Int(25)})
}

func TestCreate(t *testing.T) {
	e := newTestProc()
	e.code = []byte{
		format.CreateOp, 0, 0, 0, 0, 0,
	}
	e.classes = []format.Class{
		{Fieldc: 1},
	}

	p := &Thread{process: e}
	p.data[0] = Int(4)

	p.step()

	assert.Equal(t, p.value.Class, 0)
	assert.Equal(t, p.Field(p.value, 0), Int(4))
}

func TestRet(t *testing.T) {
	e := newTestProc()
	e.code = []byte{
		format.RetOp,
	}

	p := &Thread{process: e}
	p.data[0] = Int(1)
	p.data[1] = Int(10)

	p.step()

	assert.Equal(t, -1, p.frame)
	assert.Equal(t, 10, p.codePos)
}

func TestCall(t *testing.T) {
	e := newTestProc()
	e.code = []byte{
		format.CallOp, 0, 0, 0, 0, 0,
	}
	e.bindings = []format.Implementation{
		{EntryPoint: 10},
	}
	e.classes = []format.Class{{}}
	e.methods = []format.Method{{}}
	p := &Thread{process: e}
	p.value = p.process.memory.Alloc(0, nil)

	p.step()

	assert.Equal(t, p.codePos, 10)
}

func TestRun(t *testing.T) {
	e := newTestProc()
	e.code = []byte{
		format.GlobalLoadOp, 0, 0, 0, 0,
		format.CallOp, 0, 0, 0, 0, 0,

		50: format.NaturalOp, 1, 0, 0, 0,
		format.StoreOp, 3,
		format.GlobalLoadOp, 1, 0, 0, 0,
		format.CallOp, 1, 0, 0, 0, 0,

		100: format.NaturalOp, 2, 0, 0, 0,
		format.StoreOp, 2,
		format.NaturalOp, 150, 0, 0, 0,
		format.StoreOp, 3,
		format.NaturalOp, 1, 0, 0, 0,
		format.StoreOp, 4,
		format.GlobalLoadOp, 0, 0, 0, 0,
		format.CallOp, 1, 0, 0, 0, 2,
		150: format.StoreOp, 2,
		format.NaturalOp, 1, 0, 0, 0,
		format.StoreOp, 3,
		format.GlobalLoadOp, 1, 0, 0, 0,
		format.CallOp, 2, 0, 0, 0, 0,
	}
	e.classes = make([]format.Class, 2)
	e.methods = []format.Method{{Offset: 0}, {Offset: 1}, {Offset: 1}}
	e.bindings = []format.Implementation{
		{EntryPoint: 100, Kind: format.StandardBinding},
		{EntryPoint: 50, Kind: format.StandardBinding},
		{Class: 1, EntryPoint: 0, Kind: format.ExternalBinding},
	}
	e.extern = []func(p *Thread, recv api.Object){
		func(p *Thread, recv api.Object) {
			left := AsInt(p.Arg(0))
			right := AsInt(p.Arg(1))
			p.Return(Int(left + right))
		},
	}
	e.globals = []api.Object{
		e.memory.Alloc(0, nil),
		e.memory.Alloc(1, nil),
	}
	p := &Thread{process: e}

	p.run()

	assert.Equal(t, Int(3), p.value)
}
