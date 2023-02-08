package env

import (
	"fmt"

	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/runtime/api"
	"github.com/bobappleyard/cezanne/runtime/memory"
)

type Process struct {
	callMethodID format.MethodID
	extern       []func(p *Thread, recv api.Object)
	globals      []api.Object
	classes      []format.Class
	bindings     []format.Implementation
	methods      []format.Method
	code         []byte
	memory       *memory.Arena
	processes    []Thread
}

func (e *Process) Run() {
	p := &Thread{
		process: e,
	}
	p.run()
}

func (p *Thread) Process() *Process {
	return p.process
}

// FieldCount implements memory.Env
func (e *Process) FieldCount(class format.ClassID) int {
	return int(e.classes[class].Fieldc)
}

// MarkRoots implements memory.Env
func (e *Process) MarkRoots(c *memory.Collection) {
	for i, x := range e.globals {
		e.globals[i] = c.Copy(x)
	}
	for _, p := range e.processes {
		for i := 0; i < p.frame+p.frameEnd; i++ {
			p.data[i] = c.Copy(p.data[i])
		}
	}
}

type Thread struct {
	process  *Process
	frame    int
	frameEnd int
	context  int
	codePos  int
	value    api.Object
	data     [1024]api.Object
}

func (p *Thread) run() {
	p.data[0] = Int(0)
	p.data[2] = Int(0)
	p.data[3] = Int(-1)

	p.frame = 2

	for p.codePos != -1 {
		p.step()
	}
}

func (p *Thread) step() {
	switch p.readByte() {
	case format.LoadOp:
		varID := p.readByte()

		fmt.Println("LOAD", varID)

		p.value = p.data[p.frame+varID]

	case format.StoreOp:
		varID := p.readByte()

		fmt.Println("STORE", varID)

		if varID > p.frameEnd {
			p.frameEnd = varID
		}
		p.data[p.frame+varID] = p.value

	case format.NaturalOp:
		value := p.readInt()

		fmt.Println("NATURAL", value)

		p.value = Int(value)

	case format.GlobalLoadOp:
		globalID := p.readInt()

		fmt.Println("GLOBAL_LOAD", globalID)

		p.value = p.process.globals[globalID]

	case format.GlobalStoreOp:
		globalID := p.readInt()

		fmt.Println("GLOBAL_STORE", globalID)

		p.process.globals[globalID] = p.value

	case format.CreateOp:
		classID := format.ClassID(p.readInt())
		base := p.readByte()

		fmt.Println("CREATE", classID, base)

		p.value = p.process.memory.Alloc(classID, p.data[p.frame+base:])

	case format.FieldOp:
		field := p.readInt()

		fmt.Println("FIELD", field)

		p.value = p.Field(p.value, field)

	case format.RetOp:
		fmt.Println("RETURN")

		p.ret()

	case format.CallOp:
		methodId := p.readInt()
		base := p.readByte()
		m := p.process.methods[methodId]

		fmt.Println("CALL", methodId, base, m.Name)

		p.frame += base
		p.callMethod(m)
	}
}

func (p *Thread) Arg(id int) api.Object {
	return p.data[p.frame+id+2]
}

func (p *Thread) Return(x api.Object) {
	p.value = x
	p.ret()
}

func (p *Thread) Field(x api.Object, id int) api.Object {
	return p.process.memory.Get(x, id)
}

func (p *Thread) Create(class format.ClassID, fields ...api.Object) api.Object {
	return p.process.memory.Alloc(class, fields)
}

func (p *Thread) TailCall(object api.Object, method format.MethodID, args ...api.Object) {
	for i, x := range args {
		p.data[p.frame+i+2] = x
	}
	p.value = object
	p.callMethod(p.process.methods[method])
}

func (p *Thread) readByte() int {
	res := p.process.code[p.codePos]
	p.codePos++
	return int(res)
}

func (p *Thread) readInt() int {
	b1 := p.readByte()
	b2 := p.readByte()
	b3 := p.readByte()
	b4 := p.readByte()

	return b1 | b2<<8 | b3<<16 | b4<<24
}

func (p *Thread) ret() {
	depth := AsInt(p.data[p.frame])
	codePos := AsInt(p.data[p.frame+1])

	// reset the context, if required
	if p.frame == p.context+2 {
		p.context -= AsInt(p.data[p.context])
	}

	p.frame -= depth
	p.frameEnd = depth
	p.codePos = codePos
}

func (p *Thread) callMethod(m format.Method) {
	impl := p.getMethod(p.value, m.Offset)
	if impl == nil {
		panic("unable to call method " + m.Name)
	}
	p.enterMethod(impl)
}

func (p *Thread) getMethod(object api.Object, offset int32) *format.Implementation {
	idx := int(object.Class) + int(offset)
	if idx < 0 || idx >= len(p.process.bindings) {
		return nil
	}
	if p.process.bindings[idx].Class != object.Class {
		return nil
	}
	return &p.process.bindings[idx]
}

func (p *Thread) enterMethod(method *format.Implementation) {
	if method.Kind == format.ExternalBinding {
		fmt.Println("Calling external method", method.EntryPoint)
		p.process.extern[method.EntryPoint](p, p.value)
	} else {
		fmt.Println("Moving to codepos", method.EntryPoint)
		p.codePos = int(method.EntryPoint)
	}
}
