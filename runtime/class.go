package runtime

import (
	"errors"
	"reflect"
)

var ErrWrongType = errors.New("wrong type")

type classObject struct {
	classID ClassID
	off     int
	members []Member
}

type classBuilder struct {
	env     *Env
	members []Member
}

func (c *classObject) CallMethod(p *Process) Action {
	if p.MethodID() == extendMethod {
		return p.Return(c.newClassBuilder(p))
	}
	return p.Error(ErrUnknownMember)
}

func (c *classObject) wrap(x reflect.Value) Object {
	return &wrapperObject{
		class: c,
		inner: x,
	}
}

func (c *classObject) unwrap(x Object) (reflect.Value, error) {
	wx, ok := x.(*wrapperObject)
	if !ok {
		return reflect.Value{}, ErrWrongType
	}

	return wx.inner, nil
}

func (c *classObject) newClassBuilder(p *Process) Object {
	members := make([]Member, len(c.members))
	copy(members, c.members)
	return p.Wrap(&classBuilder{
		env:     p.env,
		members: members,
	})
}
