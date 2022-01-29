package runtime

import "reflect"

type Process struct {
	env    *Env
	tail   bool
	value  Object
	err    error
	method MethodID
	args   [256]Object
	argc   int
}

type Action struct {
}

func (p *Process) Recv() Object {
	return p.value
}

func (p *Process) MethodID() MethodID {
	return p.method
}

func (p *Process) Argc() int {
	return p.argc
}

func (p *Process) Arg(n int) Object {
	return p.args[n]
}

func (p *Process) Call(recv Object, method MethodID, args ...Object) (Object, error) {
	p.CallTail(recv, method, args...)

	for p.tail {
		p.tail = false
		p.value.CallMethod(p)
	}

	return p.Result()
}

func (p *Process) CallTail(recv Object, method MethodID, args ...Object) Action {
	p.marshalArgs(args)
	p.tail = true
	p.value = recv
	p.method = method

	return Action{}
}

func (p *Process) Result() (Object, error) {
	return p.value, p.err
}

func (p *Process) Return(value Object) Action {
	p.value = value
	p.err = nil

	return Action{}
}

func (p *Process) Error(err error) Action {
	p.value = nil
	p.err = err

	return Action{}
}

func (p *Process) Wrap(x any) Object {
	switch x := x.(type) {
	case Object:
		return x
	default:
		return p.env.Wrap(reflect.ValueOf(x))
	}
}

func (p *Process) marshalArgs(args []Object) {
	ellipsis := indexOf(args, _ellipsis)
	if ellipsis != -1 {
		copy(p.args[ellipsis:], p.args[:])
		copy(p.args[:], args[:ellipsis])
		copy(p.args[ellipsis+p.argc:], args[ellipsis+1:])
		p.argc = len(args) + p.argc - 1
	} else {
		copy(p.args[:], args)
		p.argc = len(args)
	}
}

func indexOf(xs []Object, x Object) int {
	for i, xx := range xs {
		if xx == x {
			return i
		}
	}
	return -1
}
