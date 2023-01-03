package runtime

import (
	"testing"

	"github.com/bobappleyard/cezanne/format"
	"github.com/stretchr/testify/assert"
)

func TestEnterContext(t *testing.T) {
	e := &Env{
		code: []byte{
			createOp, 0, 0, 0, 0, 0,
			storeOp, 2,
			createOp, 1, 0, 0, 0, 0,
			storeOp, 3,
			globalOp, 0, 0, 0, 0,
			callOp, 0, 0, 0, 0, 0,
		},
		classes: []format.Class{
			{},
			{},
			{},
		},
		globals: []Object{
			&standardObject{classID: 2},
		},
		methods: []format.Binding{
			{ClassID: -1},
			{ClassID: 1, Start: -2},
			{ClassID: 2, Start: -1},
		},
		extern: []func(*Process){
			func(p *Process) {
				p.EnterContext(p.Arg(0), p.Arg(1))
			},
			func(p *Process) {
				p.Return(Int(2))
			},
		},
	}

	p := &Process{env: e}

	p.run()
	assert.Equal(t, Int(2), p.value)
}
