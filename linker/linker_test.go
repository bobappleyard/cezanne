package linker

import (
	"bytes"
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/format"
)

type mockLinkerEnv map[string]*format.Package

func (e mockLinkerEnv) LoadPackage(path string) (*format.Package, error) {
	if p, ok := e[path]; ok {
		return p, nil
	}
	return nil, ErrMissingPackage
}

func TestCircularImports(t *testing.T) {
	prog, err := Link(mockLinkerEnv{
		"main": &format.Package{
			Imports: []string{"a"},
		},
		"a": &format.Package{
			Imports: []string{"b"},
		},
		"b": &format.Package{
			Imports: []string{"a"},
		},
	}, "main")
	assert.Nil(t, prog)
	assert.Equal(t, err, ErrCircularImport)
}

func TestLink(t *testing.T) {
	prog, err := Link(mockLinkerEnv{
		"main": &format.Package{
			Name:    "main",
			Imports: []string{"core", "dep"},
		},
		"dep": &format.Package{
			Name: "dep",
		},
		"core": &format.Package{
			Name: "core",
			Classes: []format.Class{
				{
					FieldCount: 2,
					Methods: []format.Method{
						{Name: "add"},
						{Name: "sub"},
						{Name: "mul"},
						{Name: "div"},
					},
				},
				{
					FieldCount: 0,
					Methods: []format.Method{
						{Name: "add"},
						{Name: "sub"},
						{Name: "mul"},
						{Name: "div"},
						{Name: "to_float"},
					},
				},
			},
		},
	}, "main")
	assert.Nil(t, err)
	assert.Equal(t, prog.Included, []Package{
		{Name: "core", Path: "core", Offset: 1},
		{Name: "dep", Path: "dep", Offset: 3},
		{Name: "main", Path: "main", Offset: 3},
	})
	assert.Equal(t, prog.Classes, []Class{
		{ID: 0, FieldCount: 0},
		{ID: 1, FieldCount: 2},
		{ID: 2, FieldCount: 0},
	})
	assert.Equal(t, prog.Methods, []format.Method{
		{Name: "add", Offset: 0},
		{Name: "sub", Offset: 6},
		{Name: "mul", Offset: 4},
		{Name: "div", Offset: 2},
		{Name: "to_float", Offset: -2},
	})
	assert.Equal(t, prog.Implementations, []Implementation{
		{Class: 2, Method: 4, Symbol: "cz_impl_core_1_to_float"},
		{Class: 1, Method: 0, Symbol: "cz_impl_core_0_add"},
		{Class: 2, Method: 0, Symbol: "cz_impl_core_1_add"},
		{Class: 1, Method: 3, Symbol: "cz_impl_core_0_div"},
		{Class: 2, Method: 3, Symbol: "cz_impl_core_1_div"},
		{Class: 1, Method: 2, Symbol: "cz_impl_core_0_mul"},
		{Class: 2, Method: 2, Symbol: "cz_impl_core_1_mul"},
		{Class: 1, Method: 1, Symbol: "cz_impl_core_0_sub"},
		{Class: 2, Method: 1, Symbol: "cz_impl_core_1_sub"},
	})
}

func TestRender(t *testing.T) {
	prog, err := Link(mockLinkerEnv{
		"test": &format.Package{
			Name: "main",
		},
	}, "test")
	assert.Nil(t, err)
	var buf bytes.Buffer
	assert.Nil(t, Render(&buf, prog))

}
