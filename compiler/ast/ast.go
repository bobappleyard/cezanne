package ast

type Module struct {
	Name  string
	Funcs []Method
}

type Expr interface {
	expr()
}

type Int struct {
	Value int
}

type Ref struct {
	Name string
}

type Create struct {
	Methods []Method
}

type Method struct {
	Name string
	Args []string
	Body Expr
}

type Invoke struct {
	Object Expr
	Name   string
	Args   []Expr
}

type Handle struct {
	In   Expr
	With []Method
}

type Trigger struct {
	Name string
	Args []Expr
}

func (Int) expr()     {}
func (Ref) expr()     {}
func (Create) expr()  {}
func (Invoke) expr()  {}
func (Handle) expr()  {}
func (Trigger) expr() {}
