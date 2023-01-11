package runtime

import (
	"testing"

	"github.com/bobappleyard/cezanne/format"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	e := &Env{
		code: []byte{
			format.LoadOp, 0,
		},
	}
	p := &Process{env: e}
	p.data[0] = Int(25)

	p.step()

	assert.Equal(t, Int(25), p.value)
}

func TestStore(t *testing.T) {
	e := &Env{
		code: []byte{
			format.StoreOp, 0,
		},
	}
	p := &Process{env: e}
	p.value = Int(25)

	p.step()

	assert.Equal(t, Int(25), p.data[0])
}

func TestNatural(t *testing.T) {
	e := &Env{
		code: []byte{
			format.NaturalOp, 25, 0, 0, 0,
		},
	}
	p := &Process{env: e}

	p.step()

	assert.Equal(t, Int(25), p.value)
}

func TestBuffer(t *testing.T) {
	e := &Env{
		code: []byte{
			format.BufferOp, 9, 0, 0, 0, 14, 0, 0, 0,
			'h', 'e', 'l', 'l', 'o',
		},
	}
	p := &Process{env: e}

	p.step()

	assert.Equal(t, &bufferObject{[]byte("hello")}, p.value)
}

func TestGlobal(t *testing.T) {
	e := &Env{
		code: []byte{
			format.GlobalOp, 0, 0, 0, 0,
		},
		globals: []Object{
			Int(5),
		},
	}
	p := &Process{env: e}

	p.step()

	assert.Equal(t, Int(5), p.value)
}

func TestCreate(t *testing.T) {
	e := &Env{
		code: []byte{
			format.CreateOp, 0, 0, 0, 0, 0,
		},
		classes: []format.Class{
			{Fieldc: 1},
		},
	}
	p := &Process{env: e}
	p.data[0] = Int(4)

	p.step()

	assert.Equal(t, &standardObject{
		classID: 0,
		fields:  []Object{Int(4)},
	}, p.value)
}

func TestRet(t *testing.T) {
	e := &Env{
		code: []byte{
			format.RetOp,
		},
	}
	p := &Process{env: e}
	p.data[0] = Int(1)
	p.data[1] = Int(10)

	p.step()

	assert.Equal(t, -1, p.frame)
	assert.Equal(t, 10, p.codePos)
}

func TestCall(t *testing.T) {
	e := &Env{
		code: []byte{
			format.CallOp, 0, 0, 0, 0, 0,
		},
		methods: []format.Binding{
			{Kind: format.StandardBinding, EntryPoint: 10},
		},
	}
	p := &Process{env: e}
	p.value = &standardObject{}

	p.step()

	assert.Equal(t, 10, p.codePos)
}

func TestRun(t *testing.T) {
	e := &Env{
		code: []byte{
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
		},
		globals: []Object{
			&standardObject{},
			&standardObject{classID: 1},
		},
		methods: []format.Binding{
			{Kind: format.StandardBinding, EntryPoint: 100},
			{Kind: format.StandardBinding, EntryPoint: 50},
			{ClassID: 1, Kind: format.ExternalBinding, EntryPoint: 0},
		},
		extern: []func(*Process){
			func(p *Process) {
				left := p.Arg(0).(*intObject).value
				right := p.Arg(1).(*intObject).value
				p.value = &intObject{left + right}
				p.ret()
			},
		},
	}
	p := &Process{env: e}

	p.run()

	assert.Equal(t, Int(3), p.value)
}
