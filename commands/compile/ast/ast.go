package ast

import "github.com/bobappleyard/cezanne/format/symtab"

type Package struct {
	Name    symtab.Symbol
	Imports []Import
	Funcs   []Method
	Vars    []Var
}

type Import struct {
	Name symtab.Symbol
	Path string
}

type Var struct {
	Name  symtab.Symbol
	Value Expr
}

type Expr interface {
	expr()
}

type Int struct {
	Value int
}

type String struct {
	Value string
}

type Ref struct {
	Name symtab.Symbol
}

type Create struct {
	Methods []Method
}

type Method struct {
	Name symtab.Symbol
	Args []symtab.Symbol
	Body Expr
}

type Let struct {
	Name  symtab.Symbol
	Value Expr
	In    Expr
}

type Invoke struct {
	Object Expr
	Name   symtab.Symbol
	Args   []Expr
}

type Handle struct {
	In   Expr
	With []Method
}

type Trigger struct {
	Name symtab.Symbol
	Args []Expr
}

func (Int) expr()     {}
func (String) expr()  {}
func (Ref) expr()     {}
func (Create) expr()  {}
func (Invoke) expr()  {}
func (Handle) expr()  {}
func (Trigger) expr() {}
