package format

type ClassID uint32
type MethodID uint32

type Program struct {
	ExternalMethods []string
	GlobalCount     int32
	Classes         []Class
	MethodOffsets   []int32
	Implmentations  []Implementation
	Code            []byte
}

type Package struct {
	Imports         []string
	ExternalMethods []string
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

type Class struct {
	Name   string
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
	Name       string
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
	Pos  int32
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

var opNames = []string{
	"LOAD",
	"STORE",
	"NATURAL",
	"GLOAD",
	"GSTORE",
	"CREATE",
	"FIELD",
	"RET",
	"CALL",
}

func OpName(op byte) string {
	if op > CallOp {
		return "?"
	}
	return opNames[op]
}
