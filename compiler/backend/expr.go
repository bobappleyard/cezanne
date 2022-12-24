package backend

import (
	"fmt"

	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/sexpr"
	"github.com/bobappleyard/cezanne/slices"
)

type Module struct {
	ns        Namespace
	lastClass int
	code      []sexpr.Node
}

type Context struct {
	module     *Module
	closure    []string
	moduleName string
	globals    []string
}

func (c *Context) Init() {
	c.module = new(Module)
	c.defineRuntimeTypes()
}

func (c *Context) defineRuntimeTypes() {
	resumer := c.module.ns.DefineClass([]Mapping{{
		Method: c.method("resume"),
		Impl:   "resumer:resume",
	}})
	arith := c.module.ns.DefineClass([]Mapping{{
		Method: c.method("add"),
		Impl:   "arith:add",
	}})
	sys := c.module.ns.DefineClass([]Mapping{{
		Method: c.method("print"),
		Impl:   "sys:print",
	}})
	c.module.code = append(c.module.code,
		sexpr.Call("set!", sexpr.Var("%resume-type-offset"), sexpr.Int(resumer)),
		sexpr.DefineVar("arith", sexpr.Call("vector", sexpr.Int(arith))),
		sexpr.DefineVar("sys", sexpr.Call("vector", sexpr.Int(sys))),
	)
}

func (c *Context) CompileModule(p *ast.Module) {
	d := &Context{
		module:     c.module,
		moduleName: p.Name,
		globals: slices.Map(p.Funcs, func(f ast.Method) string {
			return f.Name
		}),
	}
	exports := d.CompileExpr(ast.Create{Methods: p.Funcs})
	for i, x := range exports.Children[1:] {
		if slices.IndexOf(d.globals, x.Symbol) == -1 {
			continue
		}
		exports.Children[i+1] = sexpr.Bool(false)
	}
	c.module.code = append(
		c.module.code,
		sexpr.DefineVar("module:"+p.Name, exports),
	)
}

func (c *Context) Render() []sexpr.Node {
	table := sexpr.DefineVar("%methods", c.module.ns.Render())
	return append(c.module.code, table)
}

func (c *Context) CompileExpr(x ast.Expr) sexpr.Node {
	switch x := x.(type) {
	case ast.Int:
		return sexpr.Int(x.Value)

	case ast.Ref:
		return c.compileVar(x)

	case ast.Create:
		return c.compileCreate(x)

	case ast.Invoke:
		return c.compileInvoke(x)

	case ast.Handle:
		return c.compileHandle(x)

	case ast.Trigger:
		return c.compileTrigger(x)

	}

	panic("unsupported syntax")
}

func (c *Context) compileVar(x ast.Ref) sexpr.Node {
	for i, v := range c.closure {
		if x.Name != v {
			continue
		}
		return sexpr.Call("vector-ref", sexpr.Var("this"), sexpr.Int(i+1))
	}
	return sexpr.Var(x.Name)
}

func (c *Context) compileCreate(x ast.Create) sexpr.Node {
	id := c.module.lastClass
	c.module.lastClass++

	vs := map[string]bool{}
	freeVariablesExpr(vs, x)
	var closure []string
	for v := range vs {
		if vs[v] {
			closure = append(closure, v)
		}
	}

	d := &Context{
		module:     c.module,
		closure:    closure,
		moduleName: c.moduleName,
		globals:    c.globals,
	}

	offset := c.module.ns.DefineClass(slices.Map(x.Methods, func(m ast.Method) Mapping {
		name := fmt.Sprintf("%d:%s", id, m.Name)
		c.module.code = append(c.module.code, d.compileMethod(name, m))
		return Mapping{
			Method: c.module.ns.Method(m.Name),
			Impl:   name,
		}
	}))

	args := make([]sexpr.Node, len(closure)+1)
	args[0] = sexpr.Int(offset)
	copy(args[1:], slices.Map(closure, func(v string) sexpr.Node {
		return c.compileVar(ast.Ref{Name: v})
	}))

	return sexpr.Call("vector", args...)
}

func (c *Context) compileInvoke(x ast.Invoke) sexpr.Node {
	targ := x.Object
	id := c.method(x.Name)
	if o, ok := x.Object.(ast.Ref); ok && slices.IndexOf(c.globals, o.Name) != -1 {
		targ = ast.Ref{Name: "module:" + c.moduleName}
		id = c.method(o.Name)
	}
	return sexpr.Call("method-invoke", append([]sexpr.Node{
		c.CompileExpr(targ),
		sexpr.Int(id),
	}, slices.Map(x.Args, c.CompileExpr)...)...)
}

func (c *Context) compileMethod(name string, method ast.Method) sexpr.Node {
	body := c.CompileExpr(method.Body)
	return sexpr.DefineFunc(name, append([]string{"this"}, method.Args...), body)
}

func (c *Context) compileHandle(x ast.Handle) sexpr.Node {
	handler := c.compileCreate(ast.Create{Methods: x.With})
	prog := sexpr.Call("lambda", sexpr.List(), c.CompileExpr(x.In))
	return sexpr.Call("install-handlers", handler, prog)
}

func (c *Context) compileTrigger(x ast.Trigger) sexpr.Node {
	return sexpr.Call("trigger-effect", append([]sexpr.Node{
		sexpr.Int(c.method(x.Name)),
	}, slices.Map(x.Args, c.CompileExpr)...)...)
}

func (c *Context) method(name string) int {
	return c.module.ns.Method(name)
}
