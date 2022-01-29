package runtime

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransformMethodName(t *testing.T) {
	assert.Equal(t, "method_name", transformMethodName("MethodName"))
}

type wrapObjectTest struct {
	called bool
}

func (x *wrapObjectTest) NoArgsNoReturn() {
	x.called = true
}

func (x *wrapObjectTest) ReturnsArg(arg *wrapObjectTest) *wrapObjectTest {
	return arg
}

func (*wrapObjectTest) ManipulatesIntArgs(x, y int) int {
	return x + y
}

func TestWrapObject(t *testing.T) {
	e := New(nil)
	x := &wrapObjectTest{}
	p := e.Process()

	xo := e.Wrap(reflect.ValueOf(x))

	_, err := p.Call(xo, e.Method("no_args_no_return"))
	assert.Nil(t, err)
	assert.True(t, x.called)

	res, err := p.Call(xo, e.Method("returns_arg"), xo)
	assert.Nil(t, err)

	xx, err := e.Unwrap(reflect.TypeOf(x), res)
	assert.Nil(t, err)
	assert.Equal(t, x, xx.Interface())

	res, err = p.Call(xo, e.Method("manipulates_int_args"), FromInt(1), FromInt(2))
	assert.Nil(t, err)
	assert.Equal(t, 3, res.(*intObject).value)
}

func BenchmarkWrapObject(b *testing.B) {
	e := New(nil)
	x := &wrapObjectTest{}
	p := e.Process()

	xo := e.Wrap(reflect.ValueOf(x))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		res, err := p.Call(xo, e.Method("manipulates_int_args"), FromInt(1), FromInt(2))
		if err != nil {
			b.Fail()
		}
		if res.(*intObject).value != 3 {
			b.Fail()
		}
	}
}
