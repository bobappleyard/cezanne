package parser

import (
	"errors"

	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/compiler/text"
)

func ParseFile(m *ast.Package, src []byte) error {
	st, err := text.Parse[token, file](parseRules{}, tokenize(src))
	if err != nil {
		return err
	}

	m.Imports = st.imports

	for _, d := range st.decls {
		switch d := d.(type) {
		case funcDecl:
			m.Funcs = append(m.Funcs, interpretFunc(d))
		}
	}

	return nil
}

type parseRules struct{}

type file struct {
	imports []ast.Import
	decls   []decl
}

type decl interface {
	decl()
}

type funcDecl struct {
	name string
	args []string
	body []expr
}

type varDecl struct {
	name  string
	value expr
}

func (funcDecl) decl() {}
func (varDecl) decl()  {}

type expr interface {
	expr()
}

type intVal struct {
	Value int
}

type varRef struct {
	Name string
}

type createObject struct {
	Methods []method
}

type method struct {
	name string
	args []string
	body []expr
}

type invokeMethod struct {
	Object expr
	Name   string
	Args   []expr
}

type handleEffects struct {
	In   expr
	With []method
}

type triggerEffect struct {
	Name string
	Args []expr
}

func (intVal) expr()        {}
func (varRef) expr()        {}
func (createObject) expr()  {}
func (invokeMethod) expr()  {}
func (handleEffects) expr() {}
func (triggerEffect) expr() {}

type rest interface {
	rest()
	attach(e expr) expr
}

type call struct {
	args []expr
	r    rest
}

type nothingMore struct{}

func (call) rest()        {}
func (nothingMore) rest() {}

func (nothingMore) attach(e expr) expr {
	return e
}

func (c call) attach(e expr) expr {
	return c.r.attach(invokeMethod{
		Name:   "call",
		Object: e,
		Args:   c.args,
	})
}

func (parseRules) ParseEmptyFile() file {
	return file{}
}

func (parseRules) ParseNewline(f file, x newline) file {
	return f
}

func (parseRules) ParseDecl(f file, x decl) file {
	return file{imports: f.imports, decls: append(f.decls, x)}
}

func (parseRules) ParseImport(f file, m importKeyword, path ident) (file, error) {
	if len(f.decls) != 0 {
		return file{}, errors.New("imports must come at the top of the file")
	}
	return file{imports: append(f.imports, ast.Import{Name: path.name, Path: path.name}), decls: f.decls}, nil
}

func (parseRules) ParseFunc(
	m funcKeyword, name ident,
	gro groupOpen, args argList, grc groupClose,
	ar fnArrow, e expr,
) funcDecl {
	return funcDecl{
		name: name.name,
		args: args.args,
		body: []expr{e},
	}
}

func (parseRules) ParseVar(kw varKeyword, name ident, value expr) varDecl {
	return varDecl{
		name:  name.name,
		value: value,
	}
}

func (parseRules) ParseGroup(gro groupOpen, e expr, grc groupClose) expr {
	return e
}

func (parseRules) ParseObject(
	gro blockOpen, methods methodList, grc blockClose,
) createObject {
	return createObject{
		Methods: methods.methods,
	}
}

func (parseRules) ParseMethod(
	name ident,
	gro groupOpen, args argList, grc groupClose,
	ar fnArrow, e expr,
) method {
	return method{
		name: name.name,
		args: args.args,
		body: []expr{e},
	}
}

func (parseRules) ParseLambda(
	gro groupOpen, args argList, grc groupClose,
	ar fnArrow, e expr,
) createObject {
	return createObject{
		Methods: []method{{
			name: "call",
			args: args.args,
			body: []expr{e},
		}},
	}
}

func (parseRules) ParseInt(x intLit) intVal {
	return intVal{x.val}
}

func (parseRules) ParseVarRef(x ident, r rest) expr {
	return r.attach(varRef{x.name})
}

func (parseRules) ParseNullRest() nothingMore {
	return nothingMore{}
}

func (parseRules) ParseCall(gro groupOpen, args paramList, grc groupClose, r rest) call {
	return call{args: args.args, r: r}
}

func (parseRules) ParseMethodCall(
	obj expr, d dot, name ident,
	gro groupOpen, params paramList, grc groupClose,
) invokeMethod {
	return invokeMethod{
		Object: obj,
		Name:   name.name,
		Args:   params.args,
	}
}

func (parseRules) ParseTrigger(
	m triggerKeyword, effname ident,
	gro groupOpen, params paramList, grc groupClose,
) triggerEffect {
	return triggerEffect{
		Name: effname.name,
		Args: params.args,
	}
}

func (parseRules) ParseHandle(
	m handleKeyword, e expr,
	bo blockOpen, handlers methodList, bc blockClose,
) handleEffects {
	return handleEffects{
		In:   e,
		With: handlers.methods,
	}
}

type argList struct {
	args []string
}

type paramList struct {
	args []expr
}

type exprList struct {
	exprs []expr
}

type methodList struct {
	methods []method
}

func (parseRules) ParseNoArgs() argList {
	return argList{}
}

func (parseRules) ParseOneArg(arg ident) argList {
	return argList{args: []string{arg.name}}
}

func (parseRules) ParseManyArgs(prev argList, sep comma, arg ident) argList {
	return argList{args: append(prev.args, arg.name)}
}

func (parseRules) ParseEmptyBody() exprList {
	return exprList{}
}

func (parseRules) ParseBodySingleExpr(e expr) exprList {
	return exprList{exprs: []expr{e}}
}

func (parseRules) ParseBodyLeadingNewline(exprs exprList, nl newline, e expr) exprList {
	return exprList{exprs: append(exprs.exprs, e)}
}

func (parseRules) ParseBodyTrailingNewline(exprs exprList, nl newline) exprList {
	return exprs
}

func (parseRules) ParseEmptyMethodList() methodList {
	return methodList{}
}

func (parseRules) ParseBodyMethod(m method) methodList {
	return methodList{methods: []method{m}}
}

func (parseRules) ParseMethodsLeadingNewline(ms methodList, nl newline, m method) methodList {
	return methodList{methods: append(ms.methods, m)}
}

func (parseRules) ParseMethodsTrailingNewline(ms methodList, nl newline) methodList {
	return ms
}

func (parseRules) ParseNoParams() paramList {
	return paramList{}
}

func (parseRules) ParseOneParam(arg expr) paramList {
	return paramList{args: []expr{arg}}
}

func (parseRules) ParseManyParams(prev paramList, sep comma, arg expr) paramList {
	return paramList{args: append(prev.args, arg)}
}
