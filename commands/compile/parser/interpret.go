package parser

import (
	"fmt"

	"github.com/bobappleyard/cezanne/commands/compile/ast"
	"github.com/bobappleyard/cezanne/format/symtab"
	"github.com/bobappleyard/cezanne/util/slices"
)

type interpreter struct {
	syms *symtab.Symtab
}

func (i *interpreter) interpretFunc(d funcDecl) ast.Method {
	return ast.Method{
		Name: i.syms.SymbolID(d.name),
		Args: slices.Map(d.args, i.syms.SymbolID),
		Body: i.interpretExpr(d.body[0]),
	}
}

func (i *interpreter) interpretMethod(d method) ast.Method {
	return ast.Method{
		Name: i.syms.SymbolID(d.name),
		Args: slices.Map(d.args, i.syms.SymbolID),
		Body: i.interpretExpr(d.body[0]),
	}
}

func (i *interpreter) interpretHandler(d method) ast.Method {
	return ast.Method{
		Name: i.syms.SymbolID(d.name),
		Args: slices.Map(append([]string{"context"}, d.args...), i.syms.SymbolID),
		Body: i.interpretExpr(d.body[0]),
	}
}

func (i *interpreter) interpretExpr(e expr) ast.Expr {
	switch e := e.(type) {
	case intVal:
		return ast.Int{Value: e.Value}
	case strVal:
		return ast.String{Value: e.Value}
	case varRef:
		return ast.Ref{Name: i.syms.SymbolID(e.Name)}
	case createObject:
		return ast.Create{
			Methods: slices.Map(e.Methods, i.interpretMethod),
		}
	case invokeMethod:
		return ast.Invoke{
			Object: i.interpretExpr(e.Object),
			Name:   i.syms.SymbolID(e.Name),
			Args:   slices.Map(e.Args, i.interpretExpr),
		}
	case handleEffects:
		return ast.Handle{
			In:   i.interpretExpr(e.In),
			With: slices.Map(e.With, i.interpretHandler),
		}
	case triggerEffect:
		return ast.Trigger{
			Name: i.syms.SymbolID(e.Name),
			Args: slices.Map(e.Args, i.interpretExpr),
		}
	}
	panic(fmt.Sprintf("unrecognized %#v", e))
}
