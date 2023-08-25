package compiler

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/compiler/backend"
	"github.com/bobappleyard/cezanne/compiler/parser"
)

type runner struct {
	err error
}

func (r *runner) run(cmd string, args ...string) {
	if r.err != nil {
		return
	}
	c := exec.Command(cmd, args...)
	buf, err := c.CombinedOutput()
	if err != nil {
		fmt.Println(string(buf))
	}
	r.err = err
}

func (r *runner) writeFile(name string, data []byte) {
	if r.err != nil {
		return
	}
	r.err = os.WriteFile(name, data, 0600)
}

func (r *runner) setErr(err error) {
	if r.err != nil {
		return
	}
	r.err = err
}

func clearBuildDir(r *runner) {
	r.run("rm", "-rf", "testdata/build")
	r.run("mkdir", "testdata/build")
}

func buildTestPackage(r *runner) {
	r.run("cp", "testdata/test-link.json", "testdata/build/link.json")
	r.run("ar", "csr", "testdata/build/test.a", "testdata/build/link.json")
	r.run("gcc", "-c", "-I..", "-o", "testdata/build/test-pkg.o", "testdata/test-pkg.c")
	r.run("ar", "csr", "testdata/build/test.a", "testdata/build/test-pkg.o")
}

func buildCompiledPackage(r *runner, p ast.Package) {
	buf, pkg, err := backend.BuildPackage(p)
	r.setErr(err)

	r.writeFile("testdata/build/pkg.c", buf)
	buf, err = json.Marshal(pkg)
	r.setErr(err)
	r.writeFile("testdata/build/link.json", buf)
	r.run("gcc", "-c", "-I..", "-o", "testdata/build/pkg.o", "testdata/build/pkg.c")
	pkgFile := fmt.Sprintf("testdata/build/%s.a", p.Name)
	r.run("ar", "csr", pkgFile, "testdata/build/pkg.o")
	r.run("ar", "csr", pkgFile, "testdata/build/link.json")
}

func buildExecutable(r *runner) {
	r.run("go", "run", "../cmd/cz-link", "-i", "testdata/build", "-o", "testdata/build/link.c", "-p", "main")
	r.run("gcc", "-c", "-I..", "-o", "testdata/build/main.o", "../main.c")
	r.run("gcc", "-c", "-I..", "-o", "testdata/build/link.o", "testdata/build/link.c")
	r.run("gcc", "-o", "testdata/build/test", "testdata/build/main.o", "testdata/build/link.o", "testdata/build/main.a", "testdata/build/test.a")
}

func TestCompiler(t *testing.T) {
	var r runner
	clearBuildDir(&r)
	buildTestPackage(&r)

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

	buildCompiledPackage(&r, pkg)
	buildExecutable(&r)

	assert.Nil(t, r.err)

	out, err := exec.Command("testdata/build/test").CombinedOutput()
	assert.Nil(t, err)
	assert.Equal(t, string(out), "1\n2\n6\n24\n120\n720\n")

}
