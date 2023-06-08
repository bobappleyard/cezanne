package backend

import (
	"github.com/bobappleyard/cezanne/commands/compile/ast"
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/format/symtab"
)

func BuildPackage(syms *symtab.Symtab, pkg ast.Package) (*format.Package, error) {
	var root method
	pkgObject := interpretExpr(globalScope(syms, pkg), &root, buildRoot(pkg))
	root.steps = append(root.steps, returnStep{val: pkgObject})

	asm := assembler{syms: syms}
	asm.writePackage(root)

	return asm.dest.Package(), nil
}

func buildRoot(pkg ast.Package) ast.Expr {
	return ast.Create{
		Methods: pkg.Funcs,
	}
}

func globalScope(syms *symtab.Symtab, pkg ast.Package) scope {
	vars := map[symtab.Symbol]binding{}
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
	return scope{
		syms:    syms,
		vars:    vars,
		imports: imports,
	}
}
