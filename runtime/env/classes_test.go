package env

import (
	"testing"

	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/runtime/api"
	"github.com/bobappleyard/cezanne/util/assert"
)

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

func TestString(t *testing.T) {
	e := newTestProc()
	e.classes = []format.Class{
		{Name: e.syms.SymbolID("Int"), Fieldc: 0},
		{Name: e.syms.SymbolID("String"), Fieldc: 1},
	}
	e.kinds = []format.ClassID{
		format.IntKind:    0,
		format.StringKind: 1,
	}

	s := e.String("hello")

	assert.Equal(t, e.AsString(s), "hello")
}
