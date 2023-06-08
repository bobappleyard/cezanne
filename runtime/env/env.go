package env

import (
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/format/symtab"
	"github.com/bobappleyard/cezanne/runtime/api"
	"github.com/bobappleyard/cezanne/runtime/memory"
	"github.com/bobappleyard/cezanne/util/slices"
)

type Env struct {
	externalMethods map[string]func(p *Thread, recv api.Object)
	heapSize        int
}

func (e *Env) Run(syms *symtab.Symtab, prog *format.Program) {
	p := &Process{
		globals: make([]api.Object, prog.GlobalCount),
		extern: slices.Map(prog.ExternalMethods, func(n symtab.Symbol) func(p *Thread, recv api.Object) {
			return e.externalMethods[syms.SymbolName(n)]
		}),
		classes:  prog.Classes,
		kinds:    prog.CoreKinds,
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
