package linker

import (
	"encoding/json"
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/must"
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
					FieldCount: 0,
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
	t.Log(string(must.Be(json.Marshal(prog))))
	// t.Fail()
}

func TestRender(t *testing.T) {
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
					FieldCount: 0,
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
	t.Log(string(Render(prog)))
	t.Fail()
}
