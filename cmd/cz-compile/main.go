package main

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"

	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/compiler/backend"
	"github.com/bobappleyard/cezanne/compiler/parser"
	"github.com/bobappleyard/cezanne/must"
)

var (
	inputDirectory  = flag.String("input", "", "source code directory")
	outputDirectory = flag.String("output", "", "generated C code directory")
	packageName     = flag.String("name", "", "name of package")
)

func main() {
	flag.Parse()

	pkg := ast.Package{
		Name: *packageName,
	}

	for _, f := range must.Be(filepath.Glob(*inputDirectory + "/*.cz")) {
		must.Succeed(parser.ParseFile(&pkg, must.Be(os.ReadFile(f))))
	}

	b, p := must.Be2(backend.BuildPackage(pkg))

	must.Succeed(os.WriteFile(*outputDirectory+"/pkg.c", b, 0666))

	b = must.Be(json.Marshal(p))

	must.Succeed(os.WriteFile(*outputDirectory+"/link.json", b, 0666))

}
