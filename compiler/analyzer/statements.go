package analyzer

import (
	"github.com/bobappleyard/cezanne/compiler/ast/core"
	"github.com/bobappleyard/cezanne/compiler/ast/shell"
)

func (c *converter) convertStmt(s shell.Stmt) stepRef {
	switch s := s.(type) {
	case shell.VarDef:
		return c.convertVarDef(s)
	case shell.ExprStmt:
		return c.convertExpr(s.Expr)
	case shell.If:
		return c.convertIf(s)
	case shell.Return:
		return c.convertReturn(s)
	}
	return c.setErr("%T: %w", s, ErrUnsupportedSyntax)
}

func (c *converter) convertVarDef(s shell.VarDef) stepRef {
	ref := c.convertExpr(s.Value)

	for i := range c.vars {
		d := &c.vars[i]
		if d.name == s.Name {
			d.ref = ref
			return ref
		}
	}

	c.vars = append(c.vars, varDef{
		name: s.Name,
		ref:  ref,
	})

	return ref
}

func (c *converter) convertIf(s shell.If) stepRef {
	cond := c.convertExpr(s.Cond)
	branch := c.addStep(nil)

	c.convertBlock(s.Then)
	to := c.addStep(nil)
	end := c.convertBlock(s.Else)

	branch.update(core.Branch{
		Cond: cond.asVar(),
		To:   to.pos,
	})
	to.update(core.Jump{
		To: end.pos,
	})

	return end
}

func (c *converter) convertReturn(s shell.Return) stepRef {
	if e, ok := s.Value.(shell.Call); ok {
		return c.convertCall(e, true)
	}

	value := c.convertExpr(s.Value)
	return c.addStep(core.Return{
		Value: value.asVar(),
	})
}
