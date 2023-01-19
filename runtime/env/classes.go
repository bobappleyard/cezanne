package env

import (
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/runtime/api"
)

const (
	intClass format.ClassID = iota
	bufferClass
	emptyClass
	contextClass
)

func Int(x int) api.Object {
	return api.Object{
		Class: intClass,
		Data:  api.Ref(x),
	}
}

func AsInt(x api.Object) int {
	return int(x.Data)
}
