package runtime

import (
	"errors"
	"testing"

	"github.com/bobappleyard/cezanne/object/op"
	"github.com/stretchr/testify/assert"
)

func BenchmarkExec(b *testing.B) {
	code := op.NewWriter()
	code.Return(op.ReturnData{Reg: 2})

	p := new(Process)
	recv := &UserObject{}
	arg := FromInt(1)
	block := &block{
		unit: &Unit{Code: code.Bytes()},
		argc: 1,
		varc: 1,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		res, err := p.Call(block, callMethod, recv, arg)
		assert.Nil(b, err)
		assert.Equal(b, 1, res.(*intObject).value)
	}
}

func TestReturnOp(t *testing.T) {
	code := op.NewWriter()
	code.Return(op.ReturnData{Reg: 1})

	vars := []Object{
		nil, _false,
	}
	b := block{
		unit: &Unit{Code: code.Bytes()},
	}
	p := new(Process)
	b.execInstructions(p, vars)
	res, err := p.Result()

	assert.Equal(t, _false, res)
	assert.Nil(t, err)
}

func TestNaturalOp(t *testing.T) {
	code := op.NewWriter()
	code.Natural(op.NaturalData{Value: 24, Into: 0})
	code.Return(op.ReturnData{Reg: 0})

	vars := []Object{
		nil,
	}
	b := block{
		unit: &Unit{Code: code.Bytes()},
	}
	p := new(Process)
	b.execInstructions(p, vars)
	res, err := p.Result()

	assert.Equal(t, 24, res.(*intObject).value)
	assert.Nil(t, err)
}

func TestLocalOp(t *testing.T) {
	code := op.NewWriter()
	code.Local(op.LocalData{Source: 1, Into: 0})
	code.Return(op.ReturnData{Reg: 0})

	vars := []Object{
		nil, _true,
	}
	b := block{
		unit: &Unit{Code: code.Bytes()},
	}
	p := new(Process)
	b.execInstructions(p, vars)
	res, err := p.Result()

	assert.Equal(t, _true, res)
	assert.Nil(t, err)
}

func TestJumpOp(t *testing.T) {
	code := op.NewWriter()
	code.Jump(op.JumpData{To: 5})
	code.Local(op.LocalData{Source: 2, Into: 1})
	code.Return(op.ReturnData{Reg: 1})

	vars := []Object{
		nil, _false, _true,
	}
	b := block{
		unit: &Unit{Code: code.Bytes()},
	}
	p := new(Process)
	b.execInstructions(p, vars)
	res, err := p.Result()

	assert.Equal(t, _false, res)
	assert.Nil(t, err)
}

func TestBranchTrue(t *testing.T) {
	code := op.NewWriter()
	code.Branch(op.BranchData{To: 6, IfNot: 2})
	code.Local(op.LocalData{Source: 2, Into: 1})
	code.Return(op.ReturnData{Reg: 1})

	vars := []Object{
		nil, _false, _true,
	}
	b := block{
		unit: &Unit{Code: code.Bytes()},
	}
	p := new(Process)
	b.execInstructions(p, vars)
	res, err := p.Result()

	assert.Equal(t, _true, res)
	assert.Nil(t, err)
}

func TestBranchFalse(t *testing.T) {
	code := op.NewWriter()
	code.Branch(op.BranchData{To: 6, IfNot: 2})
	code.Local(op.LocalData{Source: 2, Into: 1})
	code.Return(op.ReturnData{Reg: 1})

	vars := []Object{
		nil, _true, _false,
	}
	b := block{
		unit: &Unit{Code: code.Bytes()},
	}
	p := new(Process)
	b.execInstructions(p, vars)
	res, err := p.Result()

	assert.Equal(t, _true, res)
	assert.Nil(t, err)
}

func TestCallOp(t *testing.T) {
	code := op.NewWriter()
	code.Call(op.CallData{Into: 1, Method: 0, Argc: 1})
	code.Return(op.ReturnData{Reg: 1})

	var methodID MethodID
	var argc int
	x := simpleTestObject(func(p *Process) Action {
		methodID = p.MethodID()
		argc = p.Argc()
		return p.Return(p.Arg(0))
	})

	vars := []Object{
		nil, x, _false,
	}
	b := block{
		unit: &Unit{Code: code.Bytes(), Methods: []MethodID{25}},
	}
	p := new(Process)
	b.execInstructions(p, vars)
	res, err := p.Result()

	assert.Equal(t, _false, res)
	assert.Nil(t, err)
	assert.Equal(t, MethodID(25), methodID)
	assert.Equal(t, 1, argc)
}

func TestCallOpBubblesErrors(t *testing.T) {
	code := op.NewWriter()
	code.Call(op.CallData{Into: 1, Method: 0, Argc: 1})
	code.Return(op.ReturnData{Reg: 1})

	e := errors.New("test")
	x := simpleTestObject(func(p *Process) Action {
		return p.Error(e)
	})

	vars := []Object{
		nil, x, _false,
	}
	b := block{
		unit: &Unit{Code: code.Bytes(), Methods: []MethodID{25}},
	}
	p := new(Process)
	b.execInstructions(p, vars)
	res, err := p.Result()

	assert.Nil(t, res)
	assert.Equal(t, e, err)
}

func TestCallTailOp(t *testing.T) {
	code := op.NewWriter()
	code.CallTail(op.CallTailData{Into: 1, Method: 0, Argc: 1})

	vars := []Object{
		nil, _false, _false,
	}
	b := block{
		unit: &Unit{Code: code.Bytes(), Methods: []MethodID{25}},
	}
	p := new(Process)
	b.execInstructions(p, vars)
	res, err := p.Result()

	assert.Equal(t, _false, res)
	assert.Nil(t, err)
	assert.Equal(t, MethodID(25), p.MethodID())
	assert.Equal(t, 1, p.Argc())
}
