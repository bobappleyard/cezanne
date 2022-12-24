package format

type ClassID int32

type Program struct {
	ExternalMethods []string
	Classes         []Class
	Bindings        []Binding
	Code            []byte
}

type Package struct {
	Imports     []string
	Classes     []Class
	Methods     []Method
	Bindings    []Binding
	Relocations []Relocation
	Code        []byte
}

type Binding struct {
	MethodID int32
	ClassID  ClassID
	Start    int32
}

type Class struct {
	Name   string
	Fieldc int32
}

type Visibility int32

const (
	Public Visibility = iota
	Private
	External
)

type Method struct {
	Visibility Visibility
	Name       string
}

type RelocationKind int32

const (
	ImportRel RelocationKind = iota
	ClassRel
	MethodRel
)

type Relocation struct {
	Kind RelocationKind
	ID   int32
	Pos  int32
}
