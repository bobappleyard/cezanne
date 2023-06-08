package types

import "github.com/bobappleyard/cezanne/commands/compile/ast"

type Package struct {
	Exports Type
	Types   map[string]*Constructor
}

type Env struct {
	vars map[string]Type
	cons map[qname]*Constructor
}

type qname struct {
	pkg, sym string
}

func (e *Env) ImportPackage(p *Package, as string) {
	e.vars[as] = p.Exports
	for n, c := range p.Types {
		e.cons[qname{pkg: as, sym: n}] = c
	}
}

func (e *Env) DeclareType(name string, cons *Constructor) {
	e.cons[qname{sym: name}] = cons
}

func (e *Env) TypeOf(x ast.Expr) (Type, error) {
	return nil, nil
}
