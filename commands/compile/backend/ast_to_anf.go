package backend

import (
	"fmt"

	"github.com/bobappleyard/cezanne/commands/compile/ast"
	"github.com/bobappleyard/cezanne/format/symtab"
	"github.com/bobappleyard/cezanne/util/slices"
	"golang.org/x/exp/maps"
)

type scope struct {
	syms    *symtab.Symtab
	this    variable
	imports []string
	vars    map[symtab.Symbol]binding
}

type bindingKind int

const (
	missingBinding bindingKind = iota
	globalBinding
	globalMethodBinding
	closureBinding
	localBinding
	importBinding
)

type binding struct {
	kind   bindingKind
	offset int
}

func (s scope) lookup(name symtab.Symbol) binding {
	return s.vars[name]
}

func (s scope) enter(bound, free []symtab.Symbol) scope {
	vars := map[symtab.Symbol]binding{}
	for v, b := range s.vars {
		if b.kind != globalBinding && b.kind != importBinding && b.kind != globalMethodBinding {
			continue
		}
		vars[v] = b
	}
	for i, v := range free {
		vars[v] = binding{kind: closureBinding, offset: i}
	}
	for i, v := range bound {
		vars[v] = binding{kind: localBinding, offset: i}
	}
	vars[s.syms.SymbolID("this")] = binding{kind: localBinding, offset: len(bound)}
	return scope{
		syms:    s.syms,
		imports: s.imports,
		vars:    vars,
		this:    variable(len(bound)),
	}
}

func interpretExpr(s scope, dest *method, src ast.Expr) variable {
	switch src := src.(type) {
	case ast.Int:
		v := dest.nextVar()
		dest.steps = append(dest.steps, intStep{
			val:  src.Value,
			into: v,
		})
		return v

	case ast.String:
		v := dest.nextVar()
		dest.steps = append(dest.steps, stringStep{
			val:  src.Value,
			into: v,
		})
		return v

	case ast.Ref:
		return interpretLookup(s, dest, src.Name)

	case ast.Create:
		methods, freeVars := interpretClass(s, src.Methods)
		fields := slices.Map(freeVars, func(v symtab.Symbol) variable {
			return interpretLookup(s, dest, v)
		})
		v := dest.nextVar()
		dest.steps = append(dest.steps, createStep{
			into:    v,
			fields:  fields,
			methods: methods,
		})
		return v

	case ast.Invoke:
		params := slices.Map(src.Args, func(arg ast.Expr) variable {
			return interpretExpr(s, dest, arg)
		})

		if isGlobalMethodCall(s, src) {
			return interpretGlobalMethodCall(s, dest, src.Object.(ast.Ref), params)
		}

		object := interpretExpr(s, dest, src.Object)
		v := dest.nextVar()
		dest.steps = append(dest.steps, callStep{
			object: object,
			method: src.Name,
			params: params,
			into:   v,
		})
		return v

	default:
		panic(fmt.Sprintf("unsupported syntax: %T", src))
	}
}

func isGlobalMethodCall(s scope, src ast.Invoke) bool {
	o, ok := src.Object.(ast.Ref)
	return ok && src.Name == s.syms.SymbolID("call") && s.lookup(o.Name).kind == globalMethodBinding
}

func interpretGlobalMethodCall(s scope, dest *method, src ast.Ref, params []variable) variable {
	u := dest.nextVar()
	dest.steps = append(dest.steps, importStep{
		from: ".",
		into: u,
	})

	v := dest.nextVar()
	dest.steps = append(dest.steps, callStep{
		object: u,
		method: src.Name,
		params: params,
		into:   v,
	})

	return v
}

func interpretLookup(s scope, dest *method, name symtab.Symbol) variable {
	b := s.lookup(name)
	switch b.kind {
	case localBinding:
		return variable(b.offset)

	case importBinding:
		v := dest.nextVar()
		dest.steps = append(dest.steps, importStep{
			from: s.imports[b.offset],
			into: v,
		})
		return v

	case closureBinding:
		v := dest.nextVar()
		dest.steps = append(dest.steps, fieldStep{
			from:  s.this,
			field: b.offset,
			into:  v,
		})
		return v

	default:
		panic(fmt.Sprintf("unknown variable %s", s.syms.SymbolName(name)))

	}
}

func uniq(syms []symtab.Symbol) []symtab.Symbol {
	m := map[symtab.Symbol]bool{}
	for _, s := range syms {
		m[s] = true
	}
	return maps.Keys(m)
}

func interpretClass(s scope, methods []ast.Method) ([]method, []symtab.Symbol) {
	freevars := uniq(objectFreeVars(s, methods))

	blocks := slices.Map(methods, func(m ast.Method) method {
		res := method{
			name: m.Name,
			argc: len(m.Args),
			varc: len(m.Args) + 1,
		}
		v := interpretExpr(s.enter(m.Args, freevars), &res, m.Body)
		res.steps = append(res.steps, returnStep{val: v})
		return res
	})

	return blocks, freevars
}

func objectFreeVars(s scope, methods []ast.Method) []symtab.Symbol {
	var freeVars []symtab.Symbol
	for _, m := range methods {
		freeVars = append(freeVars, exprFreeVars(s.enter(m.Args, nil), m.Body)...)
	}
	return freeVars
}

func exprFreeVars(s scope, x ast.Expr) []symtab.Symbol {
	switch x := x.(type) {
	case ast.Int, ast.String:
		return nil

	case ast.Ref:
		if s.lookup(x.Name).kind != missingBinding {
			return nil
		}
		return []symtab.Symbol{x.Name}

	case ast.Create:
		freeVars := objectFreeVars(s, x.Methods)
		var res []symtab.Symbol
		for _, v := range freeVars {
			if s.lookup(v).kind == localBinding {
				continue
			}
			res = append(res, v)
		}
		return res

	case ast.Invoke:
		var freeVars []symtab.Symbol
		for _, x := range x.Args {
			freeVars = append(freeVars, exprFreeVars(s, x)...)
		}
		freeVars = append(freeVars, exprFreeVars(s, x.Object)...)
		return freeVars

	default:
		panic(fmt.Sprintf("unsupported syntax: %T", x))
	}
}
