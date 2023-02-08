package env

import (
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/runtime/api"
	"github.com/bobappleyard/cezanne/runtime/memory"
	"github.com/bobappleyard/cezanne/slices"
)

type Env struct {
	externalMethods map[string]func(p *Thread, recv api.Object)
	heapSize        int
}

func (e *Env) Run(prog *format.Program) {
	p := &Process{
		globals: make([]api.Object, prog.GlobalCount),
		extern: slices.Map(prog.ExternalMethods, func(n string) func(p *Thread, recv api.Object) {
			return e.externalMethods[n]
		}),
		classes:  prog.Classes,
		bindings: prog.Implmentations,
		methods:  prog.Methods,
		code:     prog.Code,
	}
	p.memory = memory.NewArena(p, e.heapSize)
	p.Run()
}

func (e *Env) SetHeapSize(size int) {
	e.heapSize = size
}

func (e *Env) AddExternalMethod(name string, impl func(p *Thread, recv api.Object)) {
	if e.externalMethods == nil {
		e.externalMethods = map[string]func(p *Thread, recv api.Object){}
	}
	e.externalMethods[name] = impl
}
