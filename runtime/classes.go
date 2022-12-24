package runtime

import "github.com/bobappleyard/cezanne/format"

type intObject struct {
	value int
}

func intValue(x int) Object {
	return &intObject{x}
}

func (*intObject) ClassID() format.ClassID {
	panic("unimplemented")
}

type methodObject struct {
	offset int
}

func (*methodObject) ClassID() format.ClassID {
	panic("unimplemented")
}

type bufferObject struct {
	data []byte
}

func (*bufferObject) ClassID() format.ClassID {
	panic("unimplemented")
}

type standardObject struct {
	classID format.ClassID
	fields  []Object
}

func (o *standardObject) ClassID() format.ClassID {
	return o.classID
}
