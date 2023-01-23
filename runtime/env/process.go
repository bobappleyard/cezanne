package env

import (
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/runtime/api"
)

type Process struct {
	env      *Env
	frame    int
	frameEnd int
	context  int
	codePos  int
	value    api.Object
	data     [1024]api.Object
}

func (p *Process) run() {
	p.data[0] = Int(0)
	p.data[2] = Int(0)
	p.data[3] = Int(-1)

	p.frame = 2

	for p.codePos != -1 {
		p.step()
	}
}

func (p *Process) step() {
	switch p.readByte() {
	case format.LoadOp:
		varID := p.readByte()
		p.value = p.data[p.frame+varID]

	case format.StoreOp:
		varID := p.readByte()
		if varID > p.frameEnd {
			p.frameEnd = varID
		}
		p.data[p.frame+varID] = p.value

	case format.NaturalOp:
		value := p.readInt()
		p.value = Int(value)

	case format.GlobalLoadOp:
		globalID := p.readInt()
		p.value = p.env.globals[globalID]

	case format.GlobalStoreOp:
		globalID := p.readInt()
		p.env.globals[globalID] = p.value

	case format.CreateOp:
		classID := format.ClassID(p.readInt())
		base := p.readByte()
		p.value = p.env.memory.Alloc(classID, p.data[p.frame+base:])

	case format.FieldOp:
		field := p.readInt()
		p.value = p.Field(p.value, field)

	case format.RetOp:
		p.ret()

	case format.CallOp:
		methodId := p.readInt()
		base := p.readByte()
		p.frame += base
		p.callMethod(p.env.offsets[methodId])
	}
}

func (p *Process) Arg(id int) api.Object {
	return p.data[p.frame+id+2]
}

func (p *Process) Return(x api.Object) {
	p.value = x
	p.ret()
}

func (p *Process) Field(x api.Object, id int) api.Object {
	return p.env.memory.Get(x, id)
}

func (p *Process) Create(class format.ClassID, fields ...api.Object) api.Object {
	return p.env.memory.Alloc(class, fields)
}

func (p *Process) readByte() int {
	res := p.env.code[p.codePos]
	p.codePos++
	return int(res)
}

func (p *Process) readInt() int {
	b1 := p.readByte()
	b2 := p.readByte()
	b3 := p.readByte()
	b4 := p.readByte()

	return b1 | b2<<8 | b3<<16 | b4<<24
}

func (p *Process) ret() {
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

func (p *Process) callMethod(offset int32) {
	p.enterMethod(p.getMethod(p.value, offset))
}

func (p *Process) getMethod(object api.Object, offset int32) *format.Implementation {
	idx := int(object.Class) + int(offset)
	if idx < 0 || idx >= len(p.env.bindings) {
		return nil
	}
	if p.env.bindings[idx].Class != object.Class {
		return nil
	}
	return &p.env.bindings[idx]
}

func (p *Process) enterMethod(method *format.Implementation) {
	if method.Kind == format.ExternalBinding {
		p.env.extern[method.EntryPoint](p)
	} else {
		p.codePos = int(method.EntryPoint)
	}
}
