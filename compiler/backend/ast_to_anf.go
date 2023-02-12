package backend

import (
	"fmt"
	"sort"

	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/slices"
)

type scope struct {
	this    variable
	imports []string
	vars    map[string]binding
}

type bindingKind int

const (
	missingBinding bindingKind = iota
	globalBinding
	closureBinding
	localBinding
	importBinding
)

type binding struct {
	kind   bindingKind
	offset int
}

func (s scope) lookup(name string) binding {
	return s.vars[name]
}

func (s scope) enter(bound, free []string) scope {
	vars := map[string]binding{}
	for v, b := range s.vars {
		if b.kind != globalBinding && b.kind != importBinding {
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
	vars["this"] = binding{kind: localBinding, offset: len(bound)}
	return scope{imports: s.imports, vars: vars, this: variable(len(bound))}
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

	case ast.Ref:
		return interpretLookup(s, dest, src.Name)

	case ast.Create:
		methods, freeVars := interpretClass(s, src.Methods)
		fields := slices.Map(freeVars, func(v string) variable {
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

func interpretLookup(s scope, dest *method, name string) variable {
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
		panic(fmt.Sprintf("unknown variable %s", name))

	}
}

func interpretClass(s scope, methods []ast.Method) ([]method, []string) {
	freevars := unique(objectFreeVars(s, methods))

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

func objectFreeVars(s scope, methods []ast.Method) []string {
	var freeVars []string
	for _, m := range methods {
		freeVars = append(freeVars, exprFreeVars(s.enter(m.Args, nil), m.Body)...)
	}
	return freeVars
}

func exprFreeVars(s scope, x ast.Expr) []string {
	switch x := x.(type) {
	case ast.Int:
		return nil

	case ast.Ref:
		if s.lookup(x.Name).kind != missingBinding {
			return nil
		}
		return []string{x.Name}

	case ast.Create:
		freeVars := objectFreeVars(s, x.Methods)
		var res []string
		for _, v := range freeVars {
			if s.lookup(v).kind == localBinding {
				continue
			}
			res = append(res, v)
		}
		return res

	case ast.Invoke:
		var freeVars []string
		for _, x := range x.Args {
			freeVars = append(freeVars, exprFreeVars(s, x)...)
		}
		freeVars = append(freeVars, exprFreeVars(s, x.Object)...)
		return freeVars

	default:
		panic(fmt.Sprintf("unsupported syntax: %T", x))
	}
}

func unique(xs []string) []string {
	if len(xs) == 0 {
		return nil
	}
	sort.Slice(xs, func(i, j int) bool {
		return xs[i] < xs[j]
	})
	res := []string{xs[0]}
	for i, x := range xs[1:] {
		if x == xs[i] {
			continue
		}
		res = append(res, x)
	}
	return res
}
