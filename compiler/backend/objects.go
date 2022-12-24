package backend

import (
	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/sexpr"
)

type Namespace struct {
	methods  map[string]int
	mappings []Mapping
}

type Mapping struct {
	Method int
	Impl   string
}

func (n *Namespace) Method(name string) int {
	if n.methods == nil {
		n.methods = map[string]int{}
	}
	if id, ok := n.methods[name]; ok {
		return id
	}
	id := len(n.methods)
	n.methods[name] = id
	return id
}

func (n *Namespace) DefineClass(defs []Mapping) int {
	for i := range n.mappings {
		if !n.offsetAvailable(defs, i) {
			continue
		}
		n.registerDefinitions(defs, i)
		return i
	}
	id := len(n.mappings)
	n.registerDefinitions(defs, id)
	return id
}

func (n *Namespace) Render() sexpr.Node {
	res := make([]sexpr.Node, len(n.mappings)*2)
	for i, m := range n.mappings {
		res[i*2] = sexpr.Int(m.Method)
		if m.Impl == "" {
			res[i*2+1] = sexpr.Bool(false)
		} else {
			res[i*2+1] = sexpr.Var(m.Impl)
		}
	}
	return sexpr.Call("vector", res...)
}

func (n *Namespace) offsetAvailable(defs []Mapping, offset int) bool {
	for _, m := range defs {
		p := m.Method + offset
		if p >= len(n.mappings) {
			continue
		}
		if n.mappings[p].Impl != "" {
			return false
		}
	}
	return true
}

func (n *Namespace) registerDefinitions(defs []Mapping, offset int) {
	for _, m := range defs {
		p := m.Method + offset
		if p >= len(n.mappings) {
			padding := make([]Mapping, p-len(n.mappings)+1)
			n.mappings = append(n.mappings, padding...)
		}
		n.mappings[p] = m
	}
}

func freeVariablesExpr(vs map[string]bool, x ast.Expr) {
	switch x := x.(type) {

	case ast.Int:

	case ast.Ref:
		if x.Name != "this" {
			vs[x.Name] = true
		}

	case ast.Create:
		for _, m := range x.Methods {
			freeVariablesMethod(vs, m)
		}

	case ast.Invoke:
		freeVariablesExpr(vs, x.Object)
		for _, x := range x.Args {
			freeVariablesExpr(vs, x)
		}

	case ast.Handle:
		freeVariablesExpr(vs, x.In)
		for _, m := range x.With {
			freeVariablesMethod(vs, m)
		}

	case ast.Trigger:
		for _, x := range x.Args {
			freeVariablesExpr(vs, x)
		}

	default:
		panic("unsupported syntax")
	}
}

func freeVariablesMethod(vs map[string]bool, x ast.Method) {
	inner := map[string]bool{}
	freeVariablesExpr(inner, x.Body)
	for _, a := range x.Args {
		inner[a] = false
	}
	for a := range inner {
		vs[a] = vs[a] || inner[a]
	}
}
