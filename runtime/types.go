package runtime

import "github.com/bobappleyard/cezanne/format"

type Object interface {
	ClassID() format.ClassID
}

type Env struct {
	extern  []func(*Process)
	globals []Object
	classes []format.Class
	methods []format.Binding
	code    []byte
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
