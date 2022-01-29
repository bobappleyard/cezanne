package runtime

type Object interface {
	CallMethod(p *Process) Action
}

type UserObject struct {
	class  *classObject
	fields []Object
}

func (o *UserObject) CallMethod(p *Process) Action {
	idx, err := p.env.space.LookupMethod(p.MethodID(), o.class.classID, o.class.off)
	if err != nil {
		return p.Error(err)
	}

	return p.CallTail(o.class.members[idx].Implementation, callMethod, o, Ellipsis())
}

func Null() Object {
	return _null
}

func Bool(x Object) bool {
	return x != _false
}

func FromBool(b bool) Object {
	if b {
		return _true
	}
	return _false
}

func Ellipsis() Object {
	return _ellipsis
}

// Basic Builtins

var _true = &trueObject{}
var _false = &falseObject{}
var _null = &nullObject{}
var _ellipsis = &ellipsisObject{}

type trueObject struct{}

type falseObject struct{}

type nullObject struct{}

type ellipsisObject struct{}

func (o *trueObject) CallMethod(p *Process) Action {
	return p.Error(ErrUnknownMethod)
}

func (o *trueObject) String() string {
	return "true"
}

func (o *falseObject) CallMethod(p *Process) Action {
	return p.Error(ErrUnknownMethod)
}

func (o *falseObject) String() string {
	return "false"
}

func (o *nullObject) CallMethod(p *Process) Action {
	return p.Error(ErrUnknownMethod)
}

func (o *nullObject) String() string {
	return "null"
}

func (o *ellipsisObject) CallMethod(p *Process) Action {
	return p.Error(ErrUnknownMethod)
}

// Integers

type intObject struct {
	value int
}

var ints [1 << 16]Object

func init() {
	for i := range ints {
		ints[i] = &intObject{i - 1<<15}
	}
}

func FromInt(x int) Object {
	if x > -1<<15 && x < 1<<15 {
		return ints[x+1<<15]
	}
	return &intObject{value: x}
}

func (o *intObject) CallMethod(p *Process) Action {
	return p.Error(ErrUnknownMethod)
}

// Strings

type stringObject struct {
	value string
}

func FromString(x string) Object {
	return &stringObject{value: x}
}

func (o *stringObject) CallMethod(p *Process) Action {
	return p.Error(ErrUnknownMethod)
}
