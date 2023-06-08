package env

import (
	"fmt"

	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/format/symtab"
	"github.com/bobappleyard/cezanne/runtime/api"
	"github.com/bobappleyard/cezanne/runtime/memory"
)

type Process struct {
	syms     *symtab.Symtab
	extern   []func(p *Thread, recv api.Object)
	globals  []api.Object
	classes  []format.Class
	kinds    []format.ClassID
	bindings []format.Implementation
	methods  []format.Method
	code     []byte
	memory   *memory.Arena
	threads  []Thread
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

func (e *Process) FieldCount(class format.ClassID) int {
	if class < 0 {
		return -int(class) - 1
	}
	return int(e.classes[class].Fieldc)
}

func (e *Process) MarkRoots(c *memory.Collection) {
	for i, x := range e.globals {
		e.globals[i] = c.Copy(x)
	}
	for _, p := range e.threads {
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
	p.data[0] = p.process.Int(0)
	p.data[2] = p.process.Int(0)
	p.data[3] = p.process.Int(-1)

	p.frame = 2

	for p.codePos != -1 {
		p.step()
	}
}

const inDebug = false

func (p *Thread) debug(op string, args ...any) {
	if inDebug {
		fmt.Println(append([]any{p.codePos, op}, args...)...)
	}
}

func (p *Thread) step() {
	switch p.readByte() {
	case format.LoadOp:
		varID := p.readByte()

		p.debug("LOAD", varID)

		p.value = p.data[p.frame+varID]

	case format.StoreOp:
		varID := p.readByte()

		p.debug("STORE", varID)

		if varID > p.frameEnd {
			p.frameEnd = varID
		}
		p.data[p.frame+varID] = p.value

	case format.NaturalOp:
		value := p.readInt()

		p.debug("NATURAL", value)

		p.value = p.process.Int(value)

	case format.GlobalLoadOp:
		globalID := p.readInt()

		p.debug("GLOBAL_LOAD", globalID)

		p.value = p.process.globals[globalID]

	case format.GlobalStoreOp:
		globalID := p.readInt()

		p.debug("GLOBAL_STORE", globalID)

		p.process.globals[globalID] = p.value

	case format.CreateOp:
		classID := format.ClassID(p.readInt())
		base := p.readByte()

		p.debug("CREATE", classID, base)

		p.value = p.process.memory.Alloc(classID, p.data[p.frame+base:])

	case format.FieldOp:
		field := p.readInt()

		p.debug("FIELD", field)

		p.value = p.process.Field(p.value, field)

	case format.RetOp:
		p.debug("RETURN")

		p.ret()

	case format.CallOp:
		methodId := p.readInt()
		base := p.readByte()
		m := p.process.methods[methodId]

		p.debug("CALL", methodId, base, m.Name)

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
	depth := p.process.AsInt(p.data[p.frame])
	codePos := p.process.AsInt(p.data[p.frame+1])

	// reset the context, if required
	if p.frame == p.context+2 {
		p.context -= p.process.AsInt(p.data[p.context])
	}

	p.frame -= depth
	p.frameEnd = depth
	p.codePos = codePos
}

func (p *Thread) callMethod(m format.Method) {
	impl := p.getMethod(p.value, m.Offset)
	if impl == nil {
		panic("unable to call method " + p.process.syms.SymbolName(m.Name))
	}
	p.enterMethod(impl)
}

func (p *Thread) getMethod(object api.Object, offset int32) *format.Implementation {
	cid := int(object.Class)
	if cid < 0 {
		cid = int(p.process.kinds[format.ArrayKind])
	}
	idx := cid + int(offset)
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
		p.process.extern[method.EntryPoint](p, p.value)
	} else {
		p.codePos = int(method.EntryPoint)
	}
}
