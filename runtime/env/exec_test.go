package env

import (
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/runtime/api"
)

func TestLoad(t *testing.T) {
	e := New(32)
	e.code = []byte{
		format.LoadOp, 0,
	}
	p := &Process{env: e}
	p.data[0] = Int(25)

	p.step()

	assert.Equal(t, p.value, Int(25))
}

func TestStore(t *testing.T) {
	e := New(32)
	e.code = []byte{
		format.StoreOp, 0,
	}
	p := &Process{env: e}
	p.value = Int(25)

	p.step()

	assert.Equal(t, p.data[0], Int(25))
}

func TestNatural(t *testing.T) {
	e := New(32)
	e.code = []byte{
		format.NaturalOp, 25, 0, 0, 0,
	}
	p := &Process{env: e}

	p.step()

	assert.Equal(t, Int(25), p.value)
}

func TestBuffer(t *testing.T) {
	e := New(32)
	e.code = []byte{
		format.BufferOp, 9, 0, 0, 0, 14, 0, 0, 0,
		'h', 'e', 'l', 'l', 'o',
	}
	e.classes = []format.Class{
		1: {Fieldc: 2},
	}
	p := &Process{env: e}

	p.step()

	assert.Equal(t, p.value.Class, bufferClass)
	start := AsInt(p.Field(p.value, 0))
	end := AsInt(p.Field(p.value, 1))
	assert.Equal(t, string(e.code[start:end]), "hello")
}

func TestGlobal(t *testing.T) {
	e := New(32)
	e.code = []byte{
		format.GlobalOp, 0, 0, 0, 0,
	}
	e.globals = []api.Object{
		Int(5),
	}

	p := &Process{env: e}

	p.step()

	assert.Equal(t, p.value, Int(5))
}

func TestCreate(t *testing.T) {
	e := New(32)
	e.code = []byte{
		format.CreateOp, 0, 0, 0, 0, 0,
	}
	e.classes = []format.Class{
		{Fieldc: 1},
	}

	p := &Process{env: e}
	p.data[0] = Int(4)

	p.step()

	assert.Equal(t, p.value.Class, 0)
	assert.Equal(t, p.Field(p.value, 0), Int(4))
}

func TestRet(t *testing.T) {
	e := New(32)
	e.code = []byte{
		format.RetOp,
	}

	p := &Process{env: e}
	p.data[0] = Int(1)
	p.data[1] = Int(10)

	p.step()

	assert.Equal(t, -1, p.frame)
	assert.Equal(t, 10, p.codePos)
}

func TestCall(t *testing.T) {
	e := New(32)
	e.code = []byte{
		format.CallOp, 0, 0, 0, 0, 0,
	}
	e.methods = []format.Binding{
		{EntryPoint: 41},
	}
	e.classes = []format.Class{{}}
	p := &Process{env: e}
	p.value = p.env.memory.Alloc(0, nil)

	p.step()

	assert.Equal(t, 10, p.codePos)
}

func TestRun(t *testing.T) {
	e := New(32)
	e.code = []byte{
		format.GlobalOp, 0, 0, 0, 0,
		format.CallOp, 0, 0, 0, 0, 0,

		50: format.NaturalOp, 1, 0, 0, 0,
		format.StoreOp, 3,
		format.GlobalOp, 1, 0, 0, 0,
		format.CallOp, 1, 0, 0, 0, 0,

		100: format.NaturalOp, 2, 0, 0, 0,
		format.StoreOp, 2,
		format.NaturalOp, 150, 0, 0, 0,
		format.StoreOp, 3,
		format.NaturalOp, 1, 0, 0, 0,
		format.StoreOp, 4,
		format.GlobalOp, 0, 0, 0, 0,
		format.CallOp, 1, 0, 0, 0, 2,
		150: format.StoreOp, 2,
		format.NaturalOp, 1, 0, 0, 0,
		format.StoreOp, 3,
		format.GlobalOp, 1, 0, 0, 0,
		format.CallOp, 1, 0, 0, 0, 0,
	}
	e.classes = make([]format.Class, 2)
	e.methods = []format.Binding{
		{EntryPoint: 401},
		{EntryPoint: 201},
		{ClassID: 1, EntryPoint: 2},
	}
	e.extern = []func(*Process){
		func(p *Process) {
			left := AsInt(p.Arg(0))
			right := AsInt(p.Arg(1))
			p.Return(Int(left + right))
		},
	}
	e.globals = []api.Object{
		e.memory.Alloc(0, nil),
		e.memory.Alloc(1, nil),
	}
	p := &Process{env: e}

	p.run()

	assert.Equal(t, Int(3), p.value)
}
