package text

import (
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/compiler/stream"
)

type testTok interface {
	testTok()
}

type intTok struct {
	value int
}

type plusTok struct {
}

func (intTok) testTok()  {}
func (plusTok) testTok() {}

type testExpr interface {
	testExpr()
}

type add struct {
	left, right testExpr
}

type intVal struct {
	value int
}

type intList struct {
	vals []int
}

func (add) testExpr()    {}
func (intVal) testExpr() {}

type ruleset struct {
}

func (ruleset) ParseExprInt(val intTok) intVal {
	return intVal(val)
}

func (ruleset) ParseExprAdd(left testExpr, op plusTok, right testExpr) add {
	return add{left: left, right: right}
}

type nullableRuleset struct {
}

func (nullableRuleset) ParseInt(left intList, val intTok) intList {
	return intList{append(left.vals, val.value)}
}

func (nullableRuleset) ParseNull() intList {
	return intList{}
}

func TestGrammar(t *testing.T) {
	toks := stream.FromSlice([]testTok{
		intTok{1},
		plusTok{},
		intTok{2},
		plusTok{},
		intTok{3},
	})

	expr, err := Parse[testTok, testExpr](ruleset{}, toks)
	assert.Nil(t, err)
	assert.Equal[testExpr](t, expr, add{
		left:  add{left: intVal{value: 1}, right: intVal{value: 2}},
		right: intVal{value: 3},
	})
}

func TestNullableGrammar(t *testing.T) {
	toks := stream.FromSlice([]testTok{
		intTok{1},
		intTok{2},
		intTok{3},
	})

	expr, err := Parse[testTok, intList](nullableRuleset{}, toks)
	assert.Nil(t, err)
	assert.Equal(t, intList{[]int{1, 2, 3}}, expr)
}
