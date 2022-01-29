package analyzer

import (
	"github.com/bobappleyard/cezanne/compiler/ast/core"
	"github.com/bobappleyard/cezanne/compiler/ast/shell"
)

func (c *converter) convertExpr(e shell.Expr) stepRef {
	switch e := e.(type) {
	case shell.IntLit:
		return c.convertIntLit(e)
	case shell.FltLit:
		return c.convertFltLit(e)
	case shell.VarRef:
		return c.convertVar(e)
	case shell.Call:
		return c.convertCall(e, false)
	}
	return c.setErr("%T: %w", e, ErrUnsupportedSyntax)
}

func (c *converter) convertIntLit(e shell.IntLit) stepRef {
	return c.addStep(core.IntLit{Value: e.Value})
}

func (c *converter) convertFltLit(e shell.FltLit) stepRef {
	return c.addStep(core.FltLit{Value: e.Value})
}

func (c *converter) convertVar(e shell.VarRef) stepRef {
	for _, v := range c.vars {
		if v.name == e.Name {
			return v.ref
		}
	}
	if ref, ok := c.searchArgs(e.Name); ok {
		return ref
	}
	panic("unknown variable " + e.Name)
}

func (c *converter) searchArgs(name string) (stepRef, bool) {
	for i, a := range c.args {
		if a == name {
			ref := c.addStep(core.ArgRef{
				ID: i,
			})
			c.vars = append(c.vars, varDef{
				name: name,
				ref:  ref,
			})
			return ref, true
		}
	}
	return c.badRef(), false
}

func (c *converter) convertCall(e shell.Call, tail bool) stepRef {
	method := e.Fn.(shell.MemberRef)
	args := make([]core.Var, len(e.Args))

	for i, a := range e.Args {
		args[i] = c.convertExpr(a).asVar()
	}
	object := c.convertExpr(method.Object).asVar()

	return c.addStep(core.Call{
		Object: object,
		Tail:   tail,
		Name:   c.resolveName(method.Name),
		Args:   args,
	})
}
