package ast

type Package struct {
	Name    string
	Imports []Import
	Funcs   []Method
	Vars    []Var
}

type Import struct {
	Name, Path string
}

type Var struct {
	Name  string
	Value Expr
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

type Let struct {
	Name  string
	Value Expr
	In    Expr
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
