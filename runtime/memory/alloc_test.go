package memory

import (
	"testing"

	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/runtime/api"
	"github.com/bobappleyard/cezanne/util/assert"
)

type testEnv struct {
	root api.Object
}

func (*testEnv) FieldCount(class format.ClassID) int {
	return int(class)
}

func (e *testEnv) MarkRoots(c *Collection) {
	e.root = c.Copy(e.root)
}

func TestAlloc(t *testing.T) {
	e := &testEnv{}
	a := NewArena(e, 8)

	x := a.Alloc(1, nil)
	a.Alloc(4, nil)
	e.root = a.Alloc(2, []api.Object{{Data: 1}, x})
	a.Set(x, 0, e.root)
	a.Alloc(2, nil)

	assert.Equal(t, a.allocated, 5)
	assert.Equal(t, a.Get(e.root, 0), api.Object{Data: 1})
	assert.Equal(t, a.Get(a.Get(e.root, 1), 0), e.root)
}
