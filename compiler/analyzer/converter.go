package analyzer

import (
	"errors"
	"fmt"

	"github.com/bobappleyard/cezanne/compiler/ast/core"
	"github.com/bobappleyard/cezanne/compiler/ast/shell"
)

var (
	ErrUnsupportedSyntax = errors.New("unsupported syntax")
	ErrUnboundVariable   = errors.New("unbound variable")
)

func ConvertMethod(u *core.Unit, m shell.Method) (int, error) {
	c := &converter{
		args:  m.Args,
		unit:  u,
		block: new(core.Block),
	}
	c.convertBlock(m.Body)
	if c.err != nil {
		return 0, c.err
	}
	u.Blocks = append(u.Blocks, *c.block)
	return len(u.Blocks) - 1, nil
}

type converter struct {
	args  []string
	vars  []varDef
	unit  *core.Unit
	block *core.Block
	err   error
}

type varDef struct {
	name string
	ref  stepRef
}

type stepRef struct {
	targ *core.Block
	pos  int
}

func (r stepRef) update(s core.Step) {
	if r.pos == -1 {
		return
	}
	r.targ.Body[r.pos-1] = s
}

func (r stepRef) asVar() core.Var {
	return core.Var(r.pos - 1)
}

func (c *converter) badRef() stepRef {
	return stepRef{c.block, -1}
}

func (c *converter) setErr(fstr string, args ...any) stepRef {
	if c.err == nil {
		c.err = fmt.Errorf(fstr, args...)
	}
	return c.badRef()
}

func (c *converter) addStep(s core.Step) stepRef {
	if c.err != nil {
		return c.badRef()
	}
	c.block.Body = append(c.block.Body, s)
	return stepRef{c.block, len(c.block.Body)}
}

func (c *converter) convertBlock(b []shell.Stmt) stepRef {
	if c.err != nil {
		return c.badRef()
	}
	vars := make([]varDef, len(c.vars))
	copy(vars, c.vars)
	d := &converter{
		unit:  c.unit,
		block: c.block,
		vars:  vars,
		args:  c.args,
	}
	for _, s := range b {
		d.convertStmt(s)
	}
	if d.err != nil {
		c.err = d.err
	}
	return stepRef{c.block, len(c.block.Body)}
}

func (c *converter) resolveName(name string) int {
	for i, n := range c.unit.Names {
		if n == name {
			return i
		}
	}
	c.unit.Names = append(c.unit.Names, name)
	return len(c.unit.Names) - 1
}
