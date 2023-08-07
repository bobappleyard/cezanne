package main

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/linker"
	"github.com/bobappleyard/cezanne/must"
	ar "github.com/mkrautz/goar"
)

var (
	includePath = flag.String("i", "", "path to source packages")
	outputPath  = flag.String("o", "", "path to store linker C file")
	packageName = flag.String("p", "", "name of main package")
)

func main() {
	flag.Parse()

	e := &env{base: *includePath}
	prog := must.Be(linker.Link(e, *packageName))

	f := must.Be(os.Create(*outputPath))
	defer f.Close()

	must.Succeed(linker.Render(f, prog))
}

type env struct {
	base string
}

func (e *env) PackagePath(path string) string {
	return filepath.Join(e.base, path+".a")
}

// LoadPackage implements linker.LinkerEnv.
func (e *env) LoadPackage(path string) (*format.Package, error) {
	f, err := os.Open(e.PackagePath(path))
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
