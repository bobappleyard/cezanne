package compiler

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/buildutil"
	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/compiler/backend"
	"github.com/bobappleyard/cezanne/compiler/parser"
	"github.com/bobappleyard/cezanne/must"
)

func buildTestPackage(s *buildutil.Space) {
	s.WriteFile("link.json", must.Be(os.ReadFile("testdata/test-link.json")))
	s.WriteFile("test-pkg.c", must.Be(os.ReadFile("testdata/test-pkg.c")))
	s.Compile("test-pkg.c")
	s.Package("test.a", []string{"link.json", "test-pkg.o"})
}

func buildCompiledPackage(s *buildutil.Space, p ast.Package) {
	buf, pkg, err := backend.BuildPackage(p)
	s.SetErr(err)

	s.WriteFile("pkg.c", buf)
	s.Compile("pkg.c")

	buf, err = json.Marshal(pkg)
	s.SetErr(err)
	s.WriteFile("link.json", buf)

	s.Package(fmt.Sprintf("%s.a", p.Name), []string{"pkg.o", "link.json"})
}

func buildExecutable(s *buildutil.Space) {
	cmd := exec.Command("go", "run", "../cmd/cz-link", "-i", s.Path(""), "-o", s.Path("test"), "-p", "main")
	out, err := cmd.CombinedOutput()
	if err != nil {
		s.SetErr(fmt.Errorf("%s\n%w", string(out), err))
	}
}

func TestCompiler(t *testing.T) {
	s := buildutil.New()
	buildTestPackage(s)

	pkg := ast.Package{
		Name: "main",
	}

	assert.Nil(t, parser.ParseFile(&pkg, []byte(`

import test

func main() -> fold(
	(n, acc) -> test.print(fac(n)),
	0, 
	pair(1, pair(2, pair(3, pair(4, pair(5, pair(6, empty()))))))
)

func fac(n) -> test.lte(n, 1).match({
	true() -> 1
	false() -> test.mul(n, fac(test.sub(n, 1))) 
})

// lists

func pair(h, t) -> { match(v) -> v.pair(h, t) }
func empty() -> { match(v) -> v.empty() }

func fold(f, i, xs) -> xs.match({
	empty() -> i 
	pair(h, t) -> fold(f, f(h, i), t) 
})
		
	`)))

	buildCompiledPackage(s, pkg)
	buildExecutable(s)

	assert.Nil(t, s.Err())

	out, err := exec.Command(s.Path("test")).CombinedOutput()
	assert.Nil(t, err)
	assert.Equal(t, string(out), "1\n2\n6\n24\n120\n720\n")

}
