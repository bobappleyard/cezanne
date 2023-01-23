package linker

import (
	"testing"

	"github.com/bobappleyard/cezanne/assert"
	"github.com/bobappleyard/cezanne/format/assembly"
	"github.com/bobappleyard/cezanne/runtime/api"
	"github.com/bobappleyard/cezanne/runtime/env"
)

func TestExec(t *testing.T) {
	var l Linker
	var b assembly.Block

	b.ImplementExternalMethod(b.Class("test"), b.Method("result"), "test:result")

	var res api.Object
	ext := map[string]func(*env.Process){
		"test:result": func(p *env.Process) {
			res = p.Arg(0)
			p.Return(env.Int(0))
		},
	}

	k1 := b.Location()

	b.Create(b.Class("test"), 0)
	b.Store(2)
	b.Natural(b.Fixed(3))
	b.Store(3)
	b.Natural(k1)
	b.Store(4)
	b.Load(2)
	b.Call(b.Method("return2"), 3)
	k1.Define()
	b.Store(4)
	b.Load(2)
	b.Store(3)
	b.Load(4)
	b.Store(2)
	b.Load(3)
	b.Call(b.Method("result"), 0)

	b.ImplementMethod(b.Class("test"), b.Method("return2"))
	b.Natural(b.Fixed(2))
	b.Return()

	pkg := b.Package()
	t.Log(pkg)
	l.AddPackage(pkg)
	p := l.Complete()

	t.Log(p)

	e := env.New(32)
	e.Load(ext, p)

	e.Run()

	assert.Equal(t, res, env.Int(2))
}
