package main

import (
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/format"
)

func TestEnv(t *testing.T) {
	e := env{base: "testdata"}
	pkg, err := e.LoadPackage("test")
	assert.Nil(t, err)
	assert.Equal(t, pkg, &format.Package{
		Name:    "test",
		Imports: []string{},
		Methods: []format.Method{
			{Name: "a"},
		},
		Classes: []format.Class{
			{
				Methods: []format.Method{
					{Name: "a"},
				},
			},
			{
				Methods: []format.Method{
					{Name: "a"},
				},
			},
		},
	})
}
