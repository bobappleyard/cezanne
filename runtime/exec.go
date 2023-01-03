package runtime

import "github.com/bobappleyard/cezanne/format"

const (
	loadOp = iota
	storeOp
	naturalOp
	bufferOp
	globalOp
	createOp
	retOp
	callOp
)

func (p *Process) run() {
	p.data[0] = Int(0)
	p.data[1] = &standardObject{classID: p.env.emptyClassID}
	p.data[2] = Int(0)
	p.data[3] = Int(-1)
	p.frame = 2
	for p.codePos != -1 {
		p.step()
	}
}

func (p *Process) step() {
	switch p.readByte() {
	case loadOp:
		varID := p.readByte()
		p.value = p.data[p.frame+varID]

	case storeOp:
		varID := p.readByte()
		p.data[p.frame+varID] = p.value

	case naturalOp:
		value := p.readInt()
		p.value = Int(value)

	case bufferOp:
		start := p.readInt()
		end := p.readInt()
		p.value = &bufferObject{p.env.code[start:end]}

	case globalOp:
		globalID := p.readInt()
		p.value = p.env.globals[globalID]

	case createOp:
		classID := format.ClassID(p.readInt())
		base := p.readByte()
		p.value = p.createObject(classID, p.data[p.frame+base:])

	case retOp:
		p.ret()

	case callOp:
		offset := format.MethodID(p.readInt())
		base := p.readByte()
		p.frame += base
		p.callMethod(offset)
	}
}

func (p *Process) Arg(id int) Object {
	return p.data[p.frame+id+2]
}

func (p *Process) Return(x Object) {
	p.value = x
	p.ret()
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

func (p *Process) createObject(classID format.ClassID, args []Object) Object {
	fields := make([]Object, p.env.classes[classID].Fieldc)
	copy(fields, args)

	return &standardObject{
		classID: classID,
		fields:  fields,
	}
}

func (p *Process) ret() {
	depth := p.data[p.frame].(*intObject)
	codePos := p.data[p.frame+1].(*intObject)

	p.frame -= depth.value
	p.codePos = codePos.value
}

func (p *Process) callMethod(offset format.MethodID) {
	p.enterMethod(p.getMethod(p.value, offset))
}

func (p *Process) getMethod(object Object, id format.MethodID) *format.Binding {
	classID := object.ClassID()
	idx := int(classID) + int(id)
	if idx < 0 || idx >= len(p.env.methods) {
		return nil
	}
	if p.env.methods[idx].ClassID != classID {
		return nil
	}
	return &p.env.methods[idx]
}

func (p *Process) enterMethod(method *format.Binding) {
	if method.Start < 0 {
		p.env.extern[-method.Start-1](p)
	} else {
		p.codePos = int(method.Start)
	}
}
