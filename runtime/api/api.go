package api

import "github.com/bobappleyard/cezanne/format"

type Method struct {
	Name string
	Impl func(Process)
}

type Ref uintptr

type Object struct {
	Class format.ClassID
	Data  Ref
}

type Process interface {
	Arg(id int) Object
	Return(value Object)
}
