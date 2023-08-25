package main

import (
	_ "embed"

	"bytes"
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/bobappleyard/cezanne/buildutil"
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/linker"
	"github.com/bobappleyard/cezanne/must"
	ar "github.com/mkrautz/goar"
)

var (
	includePath = flag.String("i", "", "path to source packages")
	outputPath  = flag.String("o", "", "path to store executable")
	packageName = flag.String("p", "", "name of main package")
)

//go:embed data/main.c
var mainC []byte

func main() {
	flag.Parse()

	e := &env{base: *includePath}
	prog := must.Be(linker.Link(e, *packageName))

	var buf bytes.Buffer
	must.Succeed(linker.Render(&buf, prog))

	s := buildutil.New()
	s.WriteFile("link.c", buf.Bytes())
	s.Compile("link.c")
	s.WriteFile("main.c", mainC)
	s.Compile("main.c")

	files := []string{"main.o", "link.o"}
	for i := len(prog.Included) - 1; i >= 0; i-- {
		files = append(files, e.PackagePath(prog.Included[i].Path))
	}
	s.Link("exe", files)

	must.Succeed(s.Err())

	_, err := io.Copy(must.Be(os.OpenFile(*outputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0744)), must.Be(os.Open(s.Path("exe"))))
	must.Succeed(err)
}

func writeLinkFile(path string, prog *linker.Program) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return linker.Render(f, prog)
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
