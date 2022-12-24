package anf

type StepKind int

const (
	VarStep StepKind = iota
	CallStep
	CallTailStep
	ReturnStep
	FieldStep
)

type Register int

type Var struct {
	To   Register
	From Register
}

type Call struct {
	To   Register
	Obj  Register
	Name string
	Args []Register
}

type CallTail struct {
	Obj  Register
	Name string
	Args []Register
}

type Return struct {
	Value Register
}

type Field struct {
	To   Register
	From int
}

type Create struct {
	ClassID int
	Fields  []Register
}

func (s *Steps) Kind() StepKind {
	return StepKind(s.data[s.pos])
}

func (s *Steps) Var() Var {
	to := s.data[s.pos+1]
	from := s.data[s.pos+2]
	s.pos += 3
	return Var{
		To:   Register(to),
		From: Register(from),
	}
}
