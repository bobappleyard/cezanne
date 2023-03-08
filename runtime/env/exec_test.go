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
	p.kinds = make([]format.ClassID, 16)
	return p
}

func TestReadNegative(t *testing.T) {
	p := &Thread{
		process: &Process{
			code: []byte{0xff, 0xff, 0xff, 0xff},
		},
	}
	assert.Equal(t, format.ClassID(p.readInt()), -1)
}

func TestArray(t *testing.T) {
	e := newTestProc()
	e.classes = make([]format.Class, 1)
	a := e.Array([]api.Object{e.Int(1), e.Int(2)})
	t.Log(a)
	e.globals = []api.Object{a}
	t.Log(e.memory)
	e.memory.Collect()
	t.Log(e.memory)
	assert.Equal(t, e.AsArray(a), []api.Object{e.Int(1), e.Int(2)})
}

func TestLoad(t *testing.T) {
	e := newTestProc()
	e.code = []byte{
		format.LoadOp, 0,
	}
	p := &Thread{process: e}
	p.data[0] = e.Int(25)

	p.step()

	assert.Equal(t, p.value, e.Int(25))
}

func TestStore(t *testing.T) {
	e := newTestProc()
	e.code = []byte{
		format.StoreOp, 0,
	}
	p := &Thread{process: e}
	p.value = e.Int(25)

	p.step()

	assert.Equal(t, p.data[0], e.Int(25))
}

func TestNatural(t *testing.T) {
	e := newTestProc()
	e.code = []byte{
		format.NaturalOp, 25, 0, 0, 0,
	}
	p := &Thread{process: e}

	p.step()

	assert.Equal(t, e.Int(25), p.value)
}

func TestLoadGlobal(t *testing.T) {
	e := newTestProc()
	e.code = []byte{
		format.GlobalLoadOp, 0, 0, 0, 0,
	}
	e.globals = []api.Object{
		e.Int(5),
	}

	p := &Thread{process: e}

	p.step()

	assert.Equal(t, p.value, e.Int(5))
}

func TestStoreGlobal(t *testing.T) {
	e := newTestProc()
	e.code = []byte{
		format.GlobalStoreOp, 0, 0, 0, 0,
	}
	e.globals = make([]api.Object, 1)

	p := &Thread{process: e}
	p.value = e.Int(25)
	p.step()

	assert.Equal(t, e.globals, []api.Object{e.Int(25)})
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
	p.data[0] = e.Int(4)

	p.step()

	assert.Equal(t, p.value.Class, 0)
	assert.Equal(t, e.Field(p.value, 0), e.Int(4))
}

func TestRet(t *testing.T) {
	e := newTestProc()
	e.code = []byte{
		format.RetOp,
	}

	p := &Thread{process: e}
	p.data[0] = e.Int(1)
	p.data[1] = e.Int(10)

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
			left := e.AsInt(p.Arg(0))
			right := e.AsInt(p.Arg(1))
			p.Return(e.Int(left + right))
		},
	}
	e.globals = []api.Object{
		e.memory.Alloc(0, nil),
		e.memory.Alloc(1, nil),
	}
	p := &Thread{process: e}

	p.run()

	assert.Equal(t, e.Int(3), p.value)
}
