package runtime

import "github.com/bobappleyard/cezanne/format"

const (
	loadOp = iota
	storeOp
	naturalOp
	methodOp
	bufferOp
	globalOp
	createOp
	retOp
	callOp
)

func (p *Process) run(handlers Object) {
	p.data[0] = intValue(0)
	p.data[1] = handlers
	p.data[2] = intValue(0)
	p.data[3] = intValue(-1)
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
		p.value = intValue(value)

	case methodOp:
		value := p.readInt()
		p.value = &methodObject{value}

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
		offset := p.readInt()
		base := p.readByte()
		p.callMethod(offset, base)
	}
}

func (p *Process) arg(id int) Object {
	return p.data[p.frame+id+2]
}

func (p *Process) ret() {
	depth := p.data[p.frame].(*intObject)
	codePos := p.data[p.frame+1].(*intObject)

	p.frame -= depth.value
	p.codePos = codePos.value
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

func (p *Process) callMethod(offset, base int) {
	class := p.value.ClassID()
	method := p.env.methods[int(class)+offset]
	p.frame += base

	if method.Start < 0 {
		p.env.extern[-method.Start-1](p)
	} else {
		p.codePos = int(method.Start)
	}
}
