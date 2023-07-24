package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/bobappleyard/cezanne/cc"
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/linker"
	"github.com/bobappleyard/cezanne/must"
	"github.com/bobappleyard/cezanne/slices"
	ar "github.com/mkrautz/goar"
)

func main() {
	e := &env{base: "./build"}
	target := os.Args[1]
	prog := must.Be(linker.Link(e, target))
	f := must.Be(e.LinkFile())
	must.Succeed(Render(f, prog))
	must.Succeed(f.Close())
	must.Succeed(cc.Compile(filepath.Join(e.base, "link.o"), filepath.Join(e.base, "link.c"), "."))
	sources := []string{filepath.Join(e.base, "link.o"), "main.o"}
	sources = append(sources, slices.Map(prog.Included, func(x linker.Package) string {
		return filepath.Join(e.base, x.Path+".a")
	})...)
	must.Succeed(cc.Link(target, sources))
}

type env struct {
	base string
}

// LoadPackage implements linker.LinkerEnv.
func (e *env) LoadPackage(path string) (*format.Package, error) {
	f, err := os.Open(filepath.Join(e.base, path+".a"))
	if err != nil {
		return nil, err
	}
	r := ar.NewReader(f)
	for {
		h, err := r.Next()
		if err == io.EOF {
			return nil, os.ErrNotExist
		}
		if err != nil {
			return nil, err
		}
		if h.Name == "link.json" {
			break
		}
	}
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var pkg format.Package
	err = json.Unmarshal(buf, &pkg)
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (e *env) LinkFile() (*os.File, error) {
	return os.Create(filepath.Join(e.base, "link.c"))
}
