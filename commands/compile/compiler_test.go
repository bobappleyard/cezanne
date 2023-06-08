package compile

import (
	"fmt"
	"testing"

	"github.com/bobappleyard/cezanne/commands/compile/ast"
	"github.com/bobappleyard/cezanne/commands/compile/backend"
	"github.com/bobappleyard/cezanne/commands/compile/parser"
	linker "github.com/bobappleyard/cezanne/commands/link"
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/format/assembly"
	"github.com/bobappleyard/cezanne/format/symtab"
	"github.com/bobappleyard/cezanne/runtime/api"
	"github.com/bobappleyard/cezanne/runtime/env"
	"github.com/bobappleyard/cezanne/util/assert"
)

type testLinkerEnv map[string]*format.Package

// LoadPackage implements linker.LinkerEnv
func (e testLinkerEnv) LoadPackage(path string) (*format.Package, error) {
	pkg := e[path]
	if pkg == nil {
		return nil, fmt.Errorf("%s: %w", path, linker.ErrMissingPackage)
	}
	return pkg, nil
}

func TestCompiler(t *testing.T) {
	var pt ast.Package
	var syms symtab.Symtab

	err := parser.ParseFile(&syms, &pt, []byte(`

	import test

	func main() {
		test.print("hello")
	}

	`))
	assert.Nil(t, err)

	pkg, err := backend.BuildPackage(&syms, pt)
	assert.Nil(t, err)

	t.Log(pkg)

	prog, err := linker.Link(&syms, testLinkerEnv{
		"main":    pkg,
		"test":    testPkg(&syms),
		"runtime": runtimePkg(&syms),
	})
	assert.Nil(t, err)
	t.Log(prog)

	e := new(env.Env)
	e.SetHeapSize(32)

	var logged []string
	e.AddExternalMethod("test:print", func(p *env.Thread, recv api.Object) {
		logged = append(logged, p.Process().AsString(p.Arg(0)))
		p.Return(p.Process().Int(0))
	})

	e.AddExternalMethod("test:lte", func(p *env.Thread, recv api.Object) {
		a, b := p.Process().AsInt(p.Arg(0)), p.Process().AsInt(p.Arg(1))
		p.Return(p.Process().Bool(a <= b))
	})

	e.AddExternalMethod("test:sub", func(p *env.Thread, recv api.Object) {
		a, b := p.Process().AsInt(p.Arg(0)), p.Process().AsInt(p.Arg(1))
		p.Return(p.Process().Int(a - b))
	})

	e.AddExternalMethod("test:mul", func(p *env.Thread, recv api.Object) {
		a, b := p.Process().AsInt(p.Arg(0)), p.Process().AsInt(p.Arg(1))
		p.Return(p.Process().Int(a * b))
	})

	e.AddExternalMethod("runtime:string_constant", func(p *env.Thread, recv api.Object) {
		start, end := p.Process().AsInt(p.Arg(0)), p.Process().AsInt(p.Arg(1))
		p.Return(p.Process().String(string(prog.Code[start:end])))
	})

	e.Run(&syms, prog)

	assert.Equal(t, logged, []string{"hello"})
}

func testPkg(syms *symtab.Symtab) *format.Package {
	b := assembly.New(syms)

	pkgClass := b.Class(0)

	b.Create(pkgClass, 0)
	b.Return()

	trueClass := b.Class(0)
	b.ImplementMethod(trueClass, b.Method(syms.SymbolID("match")))
	b.Load(2)
	b.Call(b.Method(syms.SymbolID("true")), 0)

	falseClass := b.Class(0)
	b.ImplementMethod(falseClass, b.Method(syms.SymbolID("match")))
	b.Load(2)
	b.Call(b.Method(syms.SymbolID("false")), 0)

	b.ImplementExternalMethod(pkgClass, b.Method(syms.SymbolID("lte")), syms.SymbolID("test:lte"))
	b.ImplementExternalMethod(pkgClass, b.Method(syms.SymbolID("sub")), syms.SymbolID("test:sub"))
	b.ImplementExternalMethod(pkgClass, b.Method(syms.SymbolID("mul")), syms.SymbolID("test:mul"))
	b.ImplementExternalMethod(pkgClass, b.Method(syms.SymbolID("print")), syms.SymbolID("test:print"))

	p := b.Package()
	p.Classes[1].Kind = format.TrueKind
	p.Classes[2].Kind = format.FalseKind

	return p

}

func runtimePkg(syms *symtab.Symtab) *format.Package {
	b := assembly.New(syms)

	pkgClass := b.Class(0)

	b.Create(pkgClass, 0)
	b.Return()

	b.ImplementExternalMethod(pkgClass, b.Method(syms.SymbolID("string_constant")), syms.SymbolID("runtime:string_constant"))

	p := b.Package()
	p.Classes = append(p.Classes, format.Class{
		Name:   syms.SymbolID("String"),
		Fieldc: 1,
		Kind:   format.StringKind,
	})

	return p
}
