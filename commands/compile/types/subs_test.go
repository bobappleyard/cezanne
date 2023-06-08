package types

import (
	"testing"

	"github.com/bobappleyard/cezanne/util/assert"
)

func TestTransitive(t *testing.T) {
	a, b, c := NewVar(), NewVar(), NewVar()

	e := &Subs{}
	a.Unify(e, b)
	b.Unify(e, c)

	assert.Equal(t, e.Resolve(a), Type(c))
	t.Log(e)

	e = &Subs{}
	b.Unify(e, c)
	a.Unify(e, b)

	assert.Equal(t, e.Resolve(a), Type(c))
	t.Log(e)

	e = &Subs{}
	b.Unify(e, c)
	b.Unify(e, a)

	assert.Equal(t, e.Resolve(c), Type(a))
	t.Log(e)
}

func TestConcreteRefs(t *testing.T) {
	a, b := NewVar(), NewVar()
	Int := &Named{Cons: &Constructor{Name: "Int"}}

	e := &Subs{}
	a.Unify(e, Int)
	b.Unify(e, Int)

	err := a.Unify(e, b)
	assert.Nil(t, err)
}

func TestSupports(t *testing.T) {
	a, b := NewVar(), NewVar()
	Void := &Named{Cons: &Constructor{Name: "Void"}}
	Int := &Named{
		Cons: &Constructor{
			Name: "Int",
		},
		Args: []Type{},
	}
	Int.Cons.Methods = Shape{
		{"Add", Int, Int, Void},
	}

	e := &Subs{}
	a.Supports(e, Shape{{"Add", Int, b, Void}})
	a.Unify(e, Int)

	t.Log(e)

	assert.Equal(t, e.Resolve(b), Type(Int))

	e = &Subs{}
	a.Unify(e, b)
	a.Supports(e, Shape{{"Add", Int, b, Void}})
	a.Unify(e, Int)

	assert.Equal(t, e.Resolve(b), Type(Int))
}
