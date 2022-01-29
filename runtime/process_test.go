package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTailCall(t *testing.T) {
	p := new(Process)

	// This should be enough to trigger a stack overflow
	times := 10000000
	x := simpleTestObject(func(p *Process) Action {
		arg := p.Arg(0)
		if times > 0 {
			times--
			return p.CallTail(p.Recv(), p.MethodID(), arg)
		}
		return p.Return(arg)
	})
	y, err := p.Call(x, 0, _false)

	require.Nil(t, err)
	assert.Equal(t, _false, y)
}

type simpleTestObject func(p *Process) Action

func (o simpleTestObject) CallMethod(p *Process) Action {
	return o(p)
}

func TestEllipsis(t *testing.T) {
	for _, test := range []struct {
		name    string
		pre, in []Object
	}{
		{
			name: "Start",
			pre:  []Object{FromInt(0)},
			in:   []Object{_ellipsis, FromInt(1), FromInt(2)},
		},
		{
			name: "Middle",
			pre:  []Object{FromInt(1)},
			in:   []Object{FromInt(0), _ellipsis, FromInt(2)},
		},
		{
			name: "End",
			pre:  []Object{FromInt(2)},
			in:   []Object{FromInt(0), FromInt(1), _ellipsis},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			p := new(Process)
			called := simpleTestObject(func(p *Process) Action {
				for i := 0; i < 3; i++ {
					assert.Equal(t, i, p.Arg(i).(*intObject).value)
				}
				return p.Return(_true)
			})
			inter := simpleTestObject(func(p *Process) Action {
				return p.CallTail(called, 1, test.in...)
			})
			res, err := p.Call(inter, 1, test.pre...)
			require.Nil(t, err)
			assert.Equal(t, _true, res)
		})
	}
}

func BenchmarkCalls(b *testing.B) {
	p := new(Process)
	x := simpleTestObject(func(p *Process) Action {
		return p.Return(p.Arg(0))
	})
	a := FromInt(1)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		res, err := p.Call(x, 1, a)
		assert.Nil(b, err)
		assert.Equal(b, 1, res.(*intObject).value)
	}
}
