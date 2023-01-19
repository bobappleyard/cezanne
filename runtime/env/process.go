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

	case format.BufferOp:
		start := p.readInt()
		end := p.readInt()
		p.value = p.bufferObject(start, end)

	case format.GlobalOp:
		globalID := p.readInt()
		p.value = p.env.globals[globalID]

	case format.CreateOp:
		classID := format.ClassID(p.readInt())
		base := p.readByte()
		p.value = p.env.memory.Alloc(classID, p.data[p.frame+base:])

	case format.RetOp:
		p.ret()

	case format.CallOp:
		offset := format.MethodID(p.readInt())
		base := p.readByte()
		p.frame += base
		p.callMethod(offset)
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

func (p *Process) bufferObject(start, end int) api.Object {
	return p.Create(bufferClass, Int(start), Int(end))
}

func (p *Process) ret() {
	depth := AsInt(p.data[p.frame])
	codePos := AsInt(p.data[p.frame+1])

	// returning from the method will leave the current handler context
	if p.frame == p.context+2 {
		p.context -= AsInt(p.data[p.frame-2])
	}

	p.frame -= depth
	p.frameEnd = depth
	p.codePos = codePos
}

func (p *Process) callMethod(offset format.MethodID) {
	p.enterMethod(p.getMethod(p.value, offset))
}

func (p *Process) getMethod(object api.Object, id format.MethodID) *format.Binding {
	idx := int(object.Class) + int(id)
	if idx < 0 || idx >= len(p.env.methods) {
		return nil
	}
	if p.env.methods[idx].ClassID != object.Class {
		return nil
	}
	return &p.env.methods[idx]
}

func (p *Process) enterMethod(method *format.Binding) {
	kind := format.ImplKind(method.EntryPoint & 0b11)
	pos := method.EntryPoint >> 2
	if kind == format.ExternalBinding {
		p.env.extern[pos](p)
	} else {
		p.codePos = int(pos)
	}
}
