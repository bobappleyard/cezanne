package runtime

import (
	"reflect"
	"unicode"
	"unicode/utf8"
)

type wrappingProtocol interface {
	wrap(x reflect.Value) Object
	unwrap(x Object) (reflect.Value, error)
}

type wrapperFns struct {
	w func(x reflect.Value) Object
	u func(x Object) (reflect.Value, error)
}

// unwrap implements wrappingProtocol
func (f *wrapperFns) unwrap(x Object) (reflect.Value, error) {
	return f.u(x)
}

// wrap implements wrappingProtocol
func (f *wrapperFns) wrap(x reflect.Value) Object {
	return f.w(x)
}

type wrapperObject struct {
	class *classObject
	inner reflect.Value
}

type wrapperMethod struct {
	impl reflect.Value
	args []reflect.Type
}

func (o *wrapperObject) CallMethod(p *Process) Action {
	idx, err := p.env.space.LookupMethod(p.MethodID(), o.class.classID, o.class.off)
	if err != nil {
		return p.Error(err)
	}

	return p.CallTail(o.class.members[idx].Implementation, callMethod, o, Ellipsis())
}

func (o *wrapperMethod) CallMethod(p *Process) Action {
	if p.Argc() != len(o.args) {
		return p.Error(ErrWrongArgCount)
	}

	vars := make([]reflect.Value, p.Argc())
	err := o.bindArgs(p, vars)
	if err != nil {
		return p.Error(err)
	}

	// Unfortunately this allocates :/
	// https://github.com/golang/go/issues/49340
	res := o.impl.Call(vars)

	return o.marshalResponse(p, res)
}

var processType = reflect.TypeOf(new(Process))

func (o *wrapperMethod) bindArgs(p *Process, vars []reflect.Value) error {
	for i := 0; i < p.Argc(); i++ {
		arg, err := p.env.Unwrap(o.args[i], p.Arg(i))
		if err != nil {
			return err
		}

		vars[i] = arg
	}
	return nil
}

func (o *wrapperMethod) marshalResponse(p *Process, res []reflect.Value) Action {
	switch len(res) {
	case 0:
		return p.Return(Null())
	case 1:
		return p.Return(p.env.Wrap(res[0]))
	}
	return p.Error(ErrBadReturn)
}

func (e *Env) Wrap(x reflect.Value) Object {
	return e.wrapperFor(x.Type()).wrap(x)
}

func (e *Env) Unwrap(t reflect.Type, x Object) (reflect.Value, error) {
	return e.wrapperFor(t).unwrap(x)
}

func (e *Env) wrapperFor(t reflect.Type) wrappingProtocol {
	if c, ok := e.wrappers[t]; ok {
		return c
	}

	c := e.wrapperClassFor(t)
	e.nextClass++
	e.wrappers[t] = c

	return c
}

func (e *Env) wrapperClassFor(t reflect.Type) *classObject {
	c := &classObject{
		classID: e.nextClass,
	}
	for i := t.NumMethod() - 1; i >= 0; i-- {
		m := t.Method(i)
		if !m.IsExported() {
			continue
		}
		c.members = append(c.members, e.wrapperMethodFor(m))
	}
	c.off = e.space.Class(c.classID, extractMethodIDs(c.members))
	return c
}

func (e *Env) wrapperMethodFor(m reflect.Method) Member {
	id := e.space.Method(transformMethodName(m.Name))
	args := make([]reflect.Type, m.Type.NumIn())
	for i := m.Type.NumIn() - 1; i >= 0; i-- {
		args[i] = m.Type.In(i)
	}
	return Member{
		MethodID: id,
		Implementation: &wrapperMethod{
			impl: m.Func,
			args: args,
		},
	}
}

func extractMethodIDs(ms []Member) []MethodID {
	ids := make([]MethodID, len(ms))
	for i, m := range ms {
		ids[i] = m.MethodID
	}
	return ids
}

func transformMethodName(name string) string {
	var res []byte
	for _, c := range name {
		if !unicode.IsUpper(c) {
			res = utf8.AppendRune(res, c)
			continue
		}
		if len(res) != 0 {
			res = utf8.AppendRune(res, '_')
		}
		res = utf8.AppendRune(res, unicode.ToLower(c))
	}
	return string(res)
}
