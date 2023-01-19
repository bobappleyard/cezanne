package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/compiler/backend"
	"github.com/bobappleyard/cezanne/compiler/parser"
)

func TestCompile(t *testing.T) {
	var m ast.Module
	err := parser.ParseFile(&m, []byte(`
	
	func main() {
		sys.print(handle arith.add(1, trigger E(2)) {
			E(x) { arith.add(context.resume(x), context.resume(x)) }
		})
	}

	`))

	assert.Nil(t, err)

	m.Name = "main"

	ctx := new(backend.Context)
	ctx.Init()

	ctx.CompileModule(&m)

	prog, err := os.CreateTemp("", "")
	assert.Nil(t, err)

	err = compileApplication(ctx, prog)
	assert.Nil(t, err)

	prog.Close()

	t.Log(prog.Name())

	cmd := exec.Command("/usr/bin/chezscheme", "--script", prog.Name())

	out, err := cmd.CombinedOutput()
	assert.Nil(t, err)
	assert.Equal(t, "6\n", string(out))
}
