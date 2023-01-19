package env

import (
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/runtime/api"
	"github.com/bobappleyard/cezanne/runtime/memory"
)

type Env struct {
	callMethodID format.MethodID
	extern       []func(*Process)
	globals      []api.Object
	classes      []format.Class
	methods      []format.Binding
	code         []byte
	memory       *memory.Arena
	processes    []Process
}

func New(arenaSize int) *Env {
	e := new(Env)
	e.memory = memory.NewArena(e, arenaSize)
	return e
}

func (p *Process) Env() *Env {
	return p.env
}

func (e *Env) CommunicateLinkage(callMethodID format.MethodID) {
	e.callMethodID = callMethodID
}

func (e *Env) RegisterPackage(pkg api.Object) {
	e.globals = append(e.globals, pkg)
}

// FieldCount implements memory.Env
func (e *Env) FieldCount(class format.ClassID) int {
	return int(e.classes[class].Fieldc)
}

// MarkRoots implements memory.Env
func (e *Env) MarkRoots(c *memory.Collection) {
	for _, p := range e.processes {
		for i := 0; i < p.frame+p.frameEnd; i++ {
			p.data[i] = c.Copy(p.data[i])
		}
	}
}
