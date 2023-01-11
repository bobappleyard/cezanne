package runtime

import (
	"fmt"

	"github.com/bobappleyard/cezanne/format"
)

const (
	intClass format.ClassID = iota
	bufferClass
	emptyClass
	contextClass
)

type intObject struct {
	value int
}

func Int(x int) Object {
	return &intObject{x}
}

func AsInt(x Object) int {
	return x.(*intObject).value
}

func (x *intObject) String() string {
	return fmt.Sprint(x.value)
}

func (*intObject) ClassID() format.ClassID {
	return intClass
}

type bufferObject struct {
	data []byte
}

func (*bufferObject) ClassID() format.ClassID {
	return bufferClass
}

type standardObject struct {
	classID format.ClassID
	fields  []Object
}

func (o *standardObject) ClassID() format.ClassID {
	return o.classID
}

type contextObject struct {
	data []Object
}

func (o *contextObject) ClassID() format.ClassID {
	return contextClass
}
