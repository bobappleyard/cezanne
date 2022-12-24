package parser

import (
	"fmt"

	"github.com/bobappleyard/cezanne/compiler/ast"
)

func interpretFunc(d funcDecl) ast.Method {
	return ast.Method{
		Name: d.name,
		Args: d.args,
		Body: interpretExpr(d.body[0]),
	}
}

func interpretMethod(d method) ast.Method {
	return ast.Method{
		Name: d.name,
		Args: d.args,
		Body: interpretExpr(d.body[0]),
	}
}

func interpretHandler(d method) ast.Method {
	return ast.Method{
		Name: d.name,
		Args: append([]string{"context"}, d.args...),
		Body: interpretExpr(d.body[0]),
	}
}

func interpretExpr(e expr) ast.Expr {
	switch e := e.(type) {
	case intVal:
		return ast.Int{Value: e.Value}
	case varRef:
		return ast.Ref{Name: e.Name}
	case createObject:
		return ast.Create{
			Methods: mapSlice(e.Methods, interpretMethod),
		}
	case invokeMethod:
		return ast.Invoke{
			Object: interpretExpr(e.Object),
			Name:   e.Name,
			Args:   mapSlice(e.Args, interpretExpr),
		}
	case handleEffects:
		return ast.Handle{
			In:   interpretExpr(e.In),
			With: mapSlice(e.With, interpretHandler),
		}
	case triggerEffect:
		return ast.Trigger{
			Name: e.Name,
			Args: mapSlice(e.Args, interpretExpr),
		}
	}
	panic(fmt.Sprintf("unrecognized %#v", e))
}

func mapSlice[T, U any](xs []T, f func(T) U) []U {
	res := make([]U, len(xs))
	for i, x := range xs {
		res[i] = f(x)
	}
	return res
}
