package types

import (
	"testing"

	"github.com/bobappleyard/cezanne/util/assert"
)

func TestGenericType(t *testing.T) {
	Int := &Named{Cons: &Constructor{Name: "Int"}}

	Void := &Named{Cons: &Constructor{Name: "Void"}}

	FuncIn, FuncOut := NewVar(), NewVar()
	Function := &Constructor{
		Name: "Function",
		Args: []Type{FuncIn, FuncOut},
		Methods: Shape{
			{"call", FuncIn, FuncOut, Void},
		},
	}

	ArrayElement := NewVar()
	ArrayMapOut := NewVar()
	Array := &Constructor{
		Name: "Array",
		Args: []Type{ArrayElement},
	}
	Array.Methods = Shape{
		{"get", Int, ArrayElement, Void},
		{
			Name: "call",
			In:   &Named{Cons: Function, Args: []Type{ArrayElement, ArrayMapOut}},
			Out:  &Named{Cons: Array, Args: []Type{ArrayMapOut}},
			Eff:  Void,
		},
	}

}

func TestNamedSimpleMatches(t *testing.T) {
	Int := &Named{Cons: &Constructor{Name: "Int"}}

	e := &Subs{}
	err := Int.Unify(e, Int)

	assert.Nil(t, err)
	assert.Equal(t, Int.Apply(e), Type(Int))
}

func TestNamedGenericMatches(t *testing.T) {
	Int := &Named{Cons: &Constructor{Name: "Int"}}
	Array := &Constructor{Name: "Array", Args: []Type{NewVar()}}

	e := &Subs{}
	a, err := Array.WithArgs(e, []Type{Int})
	assert.Nil(t, err)

	b := NewVar()
	c, err := Array.WithArgs(e, []Type{b})
	assert.Nil(t, err)

	err = a.Unify(e, c)
	assert.Nil(t, err)

	assert.Equal(t, e.Resolve(b), Type(Int))
	assert.Equal(t, c.Apply(e), a)
}

func TestNamedSimpleFails(t *testing.T) {
	Int := &Named{Cons: &Constructor{Name: "Int"}}
	String := &Named{Cons: &Constructor{Name: "String"}}

	e := &Subs{}
	err := Int.Unify(e, String)

	assert.Equal(t, err, ErrWrongCons)
}

func TestNamedGenericFails(t *testing.T) {
	Int := &Named{Cons: &Constructor{Name: "Int"}}
	String := &Named{Cons: &Constructor{Name: "String"}}
	Array := &Constructor{Name: "Array", Args: []Type{NewVar()}}

	e := &Subs{}
	a, err := Array.WithArgs(e, []Type{Int})
	assert.Nil(t, err)

	c, err := Array.WithArgs(e, []Type{String})
	assert.Nil(t, err)

	err = a.Unify(e, c)

	assert.Equal(t, err, ErrWrongCons)
}
