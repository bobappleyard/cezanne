package core

type Unit struct {
	Blocks    []Block
	Names     []string
	InitBlock int
}

type Block struct {
	Argc int
	Varc int
	Body []Step
}

type Var int

type Step interface {
	step()
}

type IntLit struct {
	Value int
}

type StrLit struct {
	Value string
}

type FltLit struct {
	Value float64
}

type ArgRef struct {
	ID int
}

type BuiltinRef struct {
	ID int
}

type BlockRef struct {
	ID int
}

type Return struct {
	Value Var
}

type Branch struct {
	Cond Var
	To   int
}

type Jump struct {
	To int
}

type Call struct {
	Object Var
	Tail   bool
	Name   int
	Args   []Var
}

func (IntLit) step()     {}
func (StrLit) step()     {}
func (FltLit) step()     {}
func (ArgRef) step()     {}
func (BuiltinRef) step() {}
func (Return) step()     {}
func (Branch) step()     {}
func (Jump) step()       {}
func (Call) step()       {}
