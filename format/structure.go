package format

import "github.com/bobappleyard/cezanne/format/symtab"

type ClassID int32
type MethodID uint32

type Program struct {
	ExternalMethods []symtab.Symbol
	CoreKinds       []ClassID
	GlobalCount     int32
	Classes         []Class
	Methods         []Method
	Implmentations  []Implementation
	Symbols         symtab.Symtab
	Code            []byte
}

type Package struct {
	Imports         []string
	ExternalMethods []symtab.Symbol
	Classes         []Class
	Methods         []Method
	Implementations []Implementation
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
	EntryPoint uint32
}

type CoreKind uint32

const (
	UserKind CoreKind = iota
	IntKind
	TrueKind
	FalseKind
	ArrayKind
	StringKind

	// not a kind, but can be used to init the kind list
	AllKinds
)

type Class struct {
	Name   symtab.Symbol
	Kind   CoreKind
	Fieldc uint32
}

type Visibility int32

const (
	_ Visibility = iota
	Public
	Private
)

type Method struct {
	Visibility Visibility
	Name       symtab.Symbol
	Offset     int32
}

type RelocationKind int32

const (
	_ RelocationKind = iota
	GlobalRel
	ImportRel
	ClassRel
	MethodRel
	CodeRel
)

type Relocation struct {
	Kind RelocationKind
	ID   int32
	Pos  uint32
}

const (
	LoadOp = iota
	StoreOp
	NaturalOp
	GlobalLoadOp
	GlobalStoreOp
	CreateOp
	FieldOp
	RetOp
	CallOp
)
