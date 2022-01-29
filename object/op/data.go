package op

type Opcode int

const (
	Return Opcode = iota

	Natural
	Local

	Jump
	Branch

	Call
	CallTail
)

type Disposition int

const (
	GetProperty Disposition = iota
	SetProperty
	CallMethod
)

type Op struct {
	Opcode      Opcode
	Disposition Disposition
	Size        int
}

type ReturnData struct {
	Reg int
}

type NaturalData struct {
	Value, Into int
}

type LocalData struct {
	Source, Into int
}

type JumpData struct {
	To int
}

type BranchData struct {
	To, IfNot int
}

type CallData struct {
	Method, Into, Argc int
	Disposition        Disposition
}

type CallTailData struct {
	Method, Into, Argc int
	Disposition        Disposition
}
