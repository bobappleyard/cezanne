package backend

import (
	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/format"
)

func BuildPackage(pkg ast.Package) (*format.Package, error) {
	var root method
	pkgObject := interpretExpr(globalScope(pkg), &root, buildRoot(pkg))
	root.steps = append(root.steps, returnStep{val: pkgObject})

	var asm assembler
	asm.writePackage(root)

	return asm.dest.Package(), nil
}

func buildRoot(pkg ast.Package) ast.Expr {
	return ast.Create{
		Methods: pkg.Funcs,
	}
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
	return scope{vars: vars, imports: imports}
}
