package analyzer

import (
	"fmt"
	"sort"

	"github.com/bobappleyard/cezanne/compiler/ast/shell"
)

// removeClosures rewrites expressions so that no lambdas have any free variables
//
// a free variable of an expression is a variable that is not bound within that expression. the
// obvious case would be arguments to a function, but it also includes local variables
//
// the approach taken here adds the free variables as extra arguments to the functions, the values
// of those arguments are furnished to a closure object constructor which then passes them on when
// called. this is very similar to currying.
func removeClosures(e shell.Expr) shell.Expr {
	switch e := e.(type) {
	case shell.IntLit, shell.StrLit, shell.FltLit, shell.VarRef:
		return e

	case shell.MemberRef:
		return shell.MemberRef{
			Object: removeClosures(e.Object),
			Name:   e.Name,
		}

	case shell.Call:
		args := make([]shell.Expr, len(e.Args))
		for i, arg := range e.Args {
			args[i] = removeClosures(arg)
		}
		return shell.Call{
			Fn:   removeClosures(e.Fn),
			Args: args,
		}

	case shell.Op:
		return shell.Op{
			Name:  e.Name,
			Left:  removeClosures(e.Left),
			Right: removeClosures(e.Right),
		}

	case shell.Lambda:
		vars := map[string]bool{}
		freeVariablesExpr(vars, e)
		args := names(vars)
		if len(args) == 0 {
			// no need to make a closure
			return e
		}
		// not strictly speaking necessary but makes it more testable
		sort.Strings(args)
		params := make([]shell.Expr, len(args))
		for i, arg := range args {
			params[i] = shell.VarRef{Name: arg}
		}
		return shell.Call{
			Fn: shell.MemberRef{
				Object: shell.MemberRef{
					Object: shell.Builtins{},
					Name:   "Closure",
				},
				Name: "create",
			},
			Args: append(
				[]shell.Expr{shell.Lambda{
					Args: append(e.Args, args...),
					Body: e.Body,
				}},
				params...,
			),
		}
	}

	panic(fmt.Sprintf("unsupported syntax %T", e))
}

func names[T comparable, U any](x map[T]U) []T {
	var res []T
	for x := range x {
		res = append(res, x)
	}
	return res
}

func freeVariablesExpr(vars map[string]bool, e shell.Expr) {
	switch e := e.(type) {
	case shell.IntLit, shell.StrLit, shell.FltLit:
		// no vars referenced

	case shell.VarRef:
		vars[e.Name] = true

	case shell.Lambda:
		inner := map[string]bool{}
		freeVariablesBlock(inner, e.Body)
		for _, v := range e.Args {
			delete(inner, v)
		}
		for v := range inner {
			vars[v] = true
		}

	case shell.MemberRef:
		freeVariablesExpr(vars, e.Object)

	case shell.Call:
		freeVariablesExpr(vars, e.Fn)
		for _, arg := range e.Args {
			freeVariablesExpr(vars, arg)
		}

	case shell.Op:
		freeVariablesExpr(vars, e.Left)
		freeVariablesExpr(vars, e.Right)
	}
}

func freeVariablesStmt(vars map[string]bool, s shell.Stmt) {
	switch s := s.(type) {
	case shell.ExprStmt:
		freeVariablesExpr(vars, s.Expr)

	case shell.Return:
		freeVariablesExpr(vars, s.Value)

	case shell.If:
		freeVariablesExpr(vars, s.Cond)
		freeVariablesBlock(vars, s.Then)
		freeVariablesBlock(vars, s.Else)

	case shell.Assign:
		freeVariablesExpr(vars, s.Target)
		freeVariablesExpr(vars, s.Value)
	}
}

func freeVariablesBlock(vars map[string]bool, block []shell.Stmt) {
	for i, s := range block {
		switch s := s.(type) {
		case shell.VarDef:
			inner := map[string]bool{}
			freeVariablesBlock(inner, block[i+1:])
			delete(inner, s.Name)
			freeVariablesExpr(inner, s.Value)
			for v := range inner {
				vars[v] = true
			}
			return

		default:
			freeVariablesStmt(vars, s)
		}
	}
}
