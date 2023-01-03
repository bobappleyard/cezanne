package runtime

import "github.com/bobappleyard/cezanne/format"

type Object interface {
	ClassID() format.ClassID
}

type Env struct {
	callMethodID format.MethodID
	emptyClassID format.ClassID
	extern       []func(*Process)
	globals      []Object
	classes      []format.Class
	methods      []format.Binding
	code         []byte
}

type Process struct {
	env     *Env
	frame   int
	context int
	codePos int
	value   Object
	data    [1024]Object
}

type Continuation struct {
	data []Object
}

func (p *Process) Env() *Env {
	return p.env
}

func (e *Env) CommunicateLinkage(callMethodID format.MethodID, emptyClassID format.ClassID) {
	e.callMethodID = callMethodID
	e.emptyClassID = emptyClassID
}

func (e *Env) RegisterPackage(pkg Object) {
	e.globals = append(e.globals, pkg)
}
