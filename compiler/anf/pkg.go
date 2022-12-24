package anf

import (
	"reflect"
	"unsafe"
)

type Package struct {
	Imports []Offset
	Methods []Method
	Classes []Class
	Data    []byte
}

type ID[T any] int

type Offset int

type Method struct {
	Name Offset
	Argc int
}

type Class struct {
	Fieldc  int
	Methods []Binding
}

type Binding struct {
	Method ID[Method]
	Varc   int
	Start  Offset
}

type Steps struct {
	data []byte
	pos  int
}

func (p *Package) String(start Offset) string {
	end := start
	for int(end) <= len(p.Data) {
		if p.Data[end] == 0 {
			break
		}
		end++
	}
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&p.Data[0])) + uintptr(start),
		Len:  int(end - start),
	}))
}

func (p *Package) Steps(start Offset) *Steps {
	return &Steps{
		data: p.Data,
		pos:  int(start),
	}
}
