package format

type ClassID int32
type MethodID int32

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
	Bindings        []Binding
	Relocations     []Relocation
	Code            []byte
}

type BindingKind int32

const (
	_ BindingKind = iota
	StandardBinding
	ExternalBinding
	HandlerBinding
)

type Binding struct {
	MethodID   MethodID
	ClassID    ClassID
	Kind       BindingKind
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
