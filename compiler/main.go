package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/compiler/backend"
	"github.com/bobappleyard/cezanne/compiler/must"
	"github.com/bobappleyard/cezanne/compiler/parser"
	"github.com/bobappleyard/cezanne/sexpr"
)

//go:embed backend/runtime.ss
var runtime []byte

type options struct {
	moduleName string
	outputFile string
	inputFiles []string
}

func parseOptions() options {
	var res options
	flag.StringVar(&res.moduleName, "n", "", "name of module being compiled")
	flag.StringVar(&res.outputFile, "o", "", "output file")
	flag.Parse()
	res.inputFiles = flag.Args()
	return res
}

func main() {
	options := parseOptions()
	m := &ast.Module{
		Name: options.moduleName,
	}
	for _, f := range options.inputFiles {
		err := parser.ParseFile(m, must.Be(ioutil.ReadFile(f)))
		if err != nil {
			panic(err)
		}
	}
	ctx := new(backend.Context)
	ctx.Init()
	ctx.CompileModule(m)
	out := must.Be(os.Create(options.outputFile))
	err := compileApplication(ctx, out)
	if err != nil {
		panic(err)
	}
	out.Close()
}

func compileApplication(ctx *backend.Context, dest io.Writer) error {
	_, err := dest.Write(runtime)
	if err != nil {
		return nil
	}
	for _, decl := range ctx.Render() {
		_, err := fmt.Fprintln(dest, decl)
		if err != nil {
			return err
		}
	}
	epb := ctx.CompileExpr(ast.Invoke{Object: ast.Ref{Name: "module:main"}, Name: "main"})
	ep := sexpr.Call("run", sexpr.Call("lambda", sexpr.List(), epb))
	_, err = fmt.Fprintln(dest, ep)
	if err != nil {
		return nil
	}
	return nil
}
