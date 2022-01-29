package runtime

import (
	"bytes"
	"reflect"
	"unsafe"
)

// A Unit organises and collects pieces of code.
type Unit struct {
	Code    []byte
	Methods []MethodID

	wrapper Object
}

type position struct {
	unit   *Unit
	offset int
}

type unitWrapper struct {
	u *Unit
}

func (u *Unit) CallMethod(p *Process) Action {
	if p.MethodID() == importMethod {
		if p.Argc() != 1 {
			return p.Error(ErrWrongArgCount)
		}
		return u.importPackage(p, p.Arg(0).(*stringObject).value)
	}
	if u.wrapper == nil {
		u.wrapper = p.Wrap(&unitWrapper{u})
	}
	return u.wrapper.CallMethod(p)
}

func (u *unitWrapper) True() Object {
	return _true
}

func (u *unitWrapper) False() Object {
	return _false
}

func (u *unitWrapper) Null() Object {
	return _null
}

func (u *unitWrapper) Block(entryPoint, argc, varc int) *block {
	return &block{
		unit:       u.u,
		entryPoint: entryPoint,
		argc:       argc,
		varc:       varc,
	}
}

func (u *unitWrapper) String(data, len int) string {
	base := uintptr(unsafe.Pointer(&u.u.Code[0]))
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: base + uintptr(data),
		Len:  len,
	}))
}

// Exec links the unit into the provided environment and executes any necessary initialisation
// code. The returned object should be used as the exported definitions of the package.
func (u *Unit) exec(p *Process) (Object, error) {
	b, err := u.link(p.env)
	if err != nil {
		return nil, err
	}
	return p.Call(b, p.env.Method("call"), p.Wrap(u))
}

func (u *Unit) link(e *Env) (Object, error) {
	pos := &position{unit: u}

	if err := u.checkHeader(pos); err != nil {
		return nil, err
	}
	if err := u.readMethods(e, pos); err != nil {
		return nil, err
	}

	varc := pos.nextReg()

	return &block{
		unit:       u,
		varc:       varc,
		entryPoint: pos.offset,
	}, nil
}

func (u *Unit) importPackage(p *Process, name string) Action {
	if pkg, ok := p.env.pkgs[name]; ok {
		return p.Return(pkg)
	}
	v, err := p.env.load(name)
	if err != nil {
		return p.Error(err)
	}
	b, err := v.link(p.env)
	if err != nil {
		return p.Error(err)
	}
	return p.CallTail(b, p.env.Method("call"), v)
}

func (u *Unit) checkHeader(pos *position) error {
	if pos.nextArg(4) != 1970955008 {
		return ErrWrongMagicNumber
	}
	if pos.nextArg(4) != 1 {
		return ErrWrongUnitVersion
	}
	return nil
}

func (u *Unit) readMethods(e *Env, pos *position) error {
	count := pos.nextArg(4)
	u.Methods = make([]MethodID, count)
	for i := 0; i < count; i++ {
		u.Methods[i] = e.Method(pos.nextString())
	}
	return nil
}

func (s *position) nextReg() int {
	r := s.unit.Code[s.offset]
	s.offset++
	return int(r)
}

func (s *position) nextString() string {
	n := bytes.IndexByte(s.unit.Code[s.offset:], 0)
	if n == -1 {
		panic("invalid bytecode")
	}
	end := n + s.offset
	res := string(s.unit.Code[s.offset:end])
	s.offset = end + 1
	return res
}

func (s *position) nextArg(size int) int {
	code := s.unit.Code
	switch size {
	case 1:
		a := int(code[s.offset])
		s.offset++
		return a

	case 2:
		a := int(code[s.offset])
		a += int(code[s.offset+1]) << 8
		s.offset += 2
		return a

	case 4:
		a := int(code[s.offset])
		a += int(code[s.offset+1]) << 8
		a += int(code[s.offset+2]) << 16
		a += int(code[s.offset+3]) << 24
		s.offset += 4
		return a
	}

	panic("invalid bytecode")
}
