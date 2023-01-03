package runtime

import (
	"fmt"

	"github.com/bobappleyard/cezanne/format"
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
