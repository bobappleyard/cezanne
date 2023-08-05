package backend

import (
	"bytes"
	"fmt"

	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/format"
)

func BuildPackage(pkg ast.Package) ([]byte, *format.Package, error) {
	var asm assembler
	asm.dest.meta.Name = pkg.Name
	for _, imp := range pkg.Imports {
		asm.dest.meta.Imports = append(asm.dest.meta.Imports, imp.Path)
	}

	s := globalScope(pkg)
	for _, f := range pkg.Funcs {
		asm.writeFunction(s, f)
	}
	asm.writePackageInit()
	asm.processPending()

	var body bytes.Buffer
	fmt.Fprintln(&body, "#include <cz.h>")
	fmt.Fprintln(&body)
	fmt.Fprintf(&body, "extern const int cz_classes_%s;\n", pkg.Name)
	fmt.Fprintln(&body)

	asm.dest.prefix.WriteTo(&body)
	fmt.Fprintln(&body)

	asm.dest.code.WriteTo(&body)

	return body.Bytes(), &asm.dest.meta, nil
}

func globalScope(pkg ast.Package) scope {
	vars := map[string]binding{}
	var imports []string
	for i, v := range pkg.Imports {
		vars[v.Name] = binding{
			kind:   importBinding,
			offset: i,
		}
		imports = append(imports, v.Path)
	}
	for _, m := range pkg.Funcs {
		vars[m.Name] = binding{
			kind: globalMethodBinding,
		}
	}
	return scope{vars: vars, pkgName: pkg.Name}
}
