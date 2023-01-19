package format

type ClassID uint32
type MethodID uint32

type Program struct {
	ExternalMethods []string
	Classes         []Class
	Bindings        []Binding
	Code            []byte
}

type Package struct {
	Imports         []string
	ExternalMethods []string
	Classes         []Class
	Methods         []Method
	Bindings        []Implementation
	Relocations     []Relocation
	Code            []byte
}

type ImplKind int32

const (
	_ ImplKind = iota
	StandardBinding
	ExternalBinding
	HandlerBinding
)

type Implementation struct {
	Class      ClassID
	Method     MethodID
	Kind       ImplKind
	EntryPoint int32
}

type Binding struct {
	ClassID    ClassID
	EntryPoint int32
}

type Class struct {
	Name   string
	Fieldc int32
}

type Visibility int32

const (
	_ Visibility = iota
	Public
	Private
)

type Method struct {
	Visibility Visibility
	Name       string
}

type RelocationKind int32

const (
	_ RelocationKind = iota
	ImportRel
	ClassRel
	MethodRel
	CodeRel
)

type Relocation struct {
	Kind RelocationKind
	ID   int32
	Pos  int32
}

const (
	LoadOp = iota
	StoreOp
	NaturalOp
	BufferOp
	GlobalOp
	CreateOp
	RetOp
	CallOp
)
