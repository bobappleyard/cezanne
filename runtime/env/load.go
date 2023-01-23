package env

import (
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/runtime/api"
	"github.com/bobappleyard/cezanne/slices"
)

var externalMethods = map[string]func(*Process){}

func (e *Env) Load(ctx map[string]func(*Process), prog *format.Program) error {
	e.globals = make([]api.Object, prog.GlobalCount)
	e.extern = slices.Map(prog.ExternalMethods, func(n string) func(*Process) {
		return ctx[n]
	})
	e.classes = prog.Classes
	e.bindings = prog.Implmentations
	e.offsets = prog.MethodOffsets
	e.code = prog.Code
	return nil
}

func (e *Env) Run() {
	p := &Process{
		env: e,
	}
	p.run()
}

func RegisterExt(name string, fn func(p *Process)) {
	externalMethods[name] = fn
}
