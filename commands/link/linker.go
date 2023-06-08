package link

import (
	"errors"

	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/format/symtab"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var (
	ErrCircularImport    = errors.New("circular import")
	ErrMissingPackage    = errors.New("missing package")
	ErrMissingMainMethod = errors.New("missinng main method")
)

type LinkerEnv interface {
	LoadPackage(path string) (*format.Package, error)
}

// Given a collection of packages keyed by path, create a program by starting
// with the "main" package and taking the transitive closure of the import
// relation.
func Link(syms *symtab.Symtab, env LinkerEnv) (*format.Program, error) {
	l := &linker{
		syms:    syms,
		env:     env,
		methods: map[symtab.Symbol]*method{syms.SymbolID("call"): {}},
		imports: map[string]*importedPackage{},
	}
	l.init()
	err := l.importPackage("main")
	if err != nil {
		return nil, err
	}
	return l.complete(), nil
}

type linker struct {
	syms    *symtab.Symtab
	env     LinkerEnv
	program format.Program
	methods map[symtab.Symbol]*method
	imports map[string]*importedPackage
}

type importedPackage struct {
	order  int
	global int32
	class  format.ClassID
}

type method struct {
	id    format.MethodID
	impls []format.Implementation
}

func (l *linker) init() {
	l.program.Classes = make([]format.Class, 2)
	l.program.CoreKinds = make([]format.ClassID, format.AllKinds)
	l.program.Code = []byte{
		format.CreateOp, 1, 0, 0, 0, 0,
		format.CallOp, 0, 0, 0, 0, 0,
	}
}

func (l *linker) complete() *format.Program {
	l.program.Classes[1].Name = l.syms.SymbolID("progInit")
	call := l.syms.SymbolID("call")
	l.methods[call].impls = append(l.methods[call].impls, format.Implementation{
		Class:      1,
		Method:     0,
		Kind:       format.StandardBinding,
		EntryPoint: uint32(len(l.program.Code)),
	})
	packages := maps.Values(l.imports)
	slices.SortFunc(packages, func(left, right *importedPackage) bool {
		return left.global < right.global
	})
	for _, p := range packages {
		l.addPkgInitCode(p)
	}
	l.addMainInitCode()
	l.determineOffsets()
	for i, c := range l.program.Classes {
		if c.Kind == format.UserKind {
			continue
		}
		l.program.CoreKinds[c.Kind] = format.ClassID(i)
	}
	return &l.program
}

func (l *linker) importPackage(path string) error {
	if p, ok := l.imports[path]; ok {
		if p.global == -1 {
			return ErrCircularImport
		}
		return nil
	}

	l.imports[path] = &importedPackage{order: len(l.imports), global: -1}

	p, err := l.env.LoadPackage(path)
	if err != nil {
		return err
	}

	for _, q := range p.Imports {
		if q == "." {
			continue
		}
		err := l.importPackage(q)
		if err != nil {
			return err
		}
	}

	global := l.program.GlobalCount
	l.program.GlobalCount++

	class := format.ClassID(len(l.program.Classes))
	l.program.Classes = append(l.program.Classes, format.Class{Name: l.syms.SymbolID("PackageInit")})

	l.addPackageEntry(class)
	l.appendPackage(p, global)

	l.imports[path].global = global
	l.imports[path].class = class

	return nil
}

func (l *linker) appendPackage(p *format.Package, pkgGlob int32) {
	l.processRelocations(p, pkgGlob)
	l.processBindings(p)
	l.program.ExternalMethods = append(l.program.ExternalMethods, p.ExternalMethods...)
	l.program.Classes = append(l.program.Classes, p.Classes...)
	l.program.Code = append(l.program.Code, p.Code...)
}

func (l *linker) addPackageEntry(packageClass format.ClassID) {
	call := l.syms.SymbolID("call")
	l.methods[call].impls = append(l.methods[call].impls, format.Implementation{
		Class:      packageClass,
		Kind:       format.StandardBinding,
		EntryPoint: uint32(len(l.program.Code)),
	})
}

func (l *linker) addMainInitCode() {
	initPos := len(l.program.Code)

	l.program.Code = append(l.program.Code,
		format.GlobalLoadOp, 0, 0, 0, 0,
		format.CallOp, 0, 0, 0, 0, 0,
	)

	main := l.syms.SymbolID("main")

	writeInt32(l.program.Code[initPos+1:], int32(l.imports["main"].global))
	writeInt32(l.program.Code[initPos+6:], int32(l.methods[main].id))
}

func (l *linker) addPkgInitCode(p *importedPackage) {
	initPos := len(l.program.Code)

	l.program.Code = append(l.program.Code,
		format.NaturalOp, 2, 0, 0, 0,
		format.StoreOp, 2,
		format.NaturalOp, 0, 0, 0, 0,
		format.StoreOp, 3,
		format.CreateOp, 0, 0, 0, 0, 0,
		format.CallOp, 0, 0, 0, 0, 2,
		format.GlobalStoreOp, 0, 0, 0, 0,
	)

	writeInt32(l.program.Code[initPos+8:], int32(initPos+26))
	writeInt32(l.program.Code[initPos+15:], int32(p.class))
	writeInt32(l.program.Code[initPos+27:], p.global)
}

func (l *linker) processRelocations(p *format.Package, pkgGlob int32) {
	var glob int32 = -1
	for _, rel := range p.Relocations {
		switch rel.Kind {
		case format.ClassRel:
			rel.ID += int32(len(l.program.Classes))

		case format.ImportRel:
			id := l.importGlobal(p.Imports[rel.ID], pkgGlob)
			rel.ID = id

		case format.GlobalRel:
			if rel.ID > glob {
				glob = rel.ID
			}
			rel.ID += l.program.GlobalCount

		case format.CodeRel:
			rel.ID += int32(len(l.program.Code))

		case format.MethodRel:
			rel.ID = int32(l.method(p.Methods[rel.ID].Name).id)
		}
		writeInt32(p.Code[rel.Pos:], rel.ID)
	}
	l.program.GlobalCount += glob + 1
}

func (l *linker) importGlobal(name string, pkgGlob int32) int32 {
	if name == "." {
		return pkgGlob
	}
	return l.imports[name].global
}

func (l *linker) processBindings(p *format.Package) {
	for _, impl := range p.Implementations {
		var ep uint32
		switch impl.Kind {
		case format.StandardBinding:
			ep = impl.EntryPoint + uint32(len(l.program.Code))
		case format.ExternalBinding:
			ep = impl.EntryPoint + uint32(len(l.program.ExternalMethods))
		}
		m := l.method(p.Methods[impl.Method].Name)
		m.impls = append(m.impls, format.Implementation{
			Class:      impl.Class + format.ClassID(len(l.program.Classes)),
			Method:     m.id,
			Kind:       impl.Kind,
			EntryPoint: ep,
		})
	}
}

func (l *linker) method(name symtab.Symbol) *method {
	if m, ok := l.methods[name]; ok {
		return m
	}
	m := &method{id: format.MethodID(len(l.methods))}
	l.methods[name] = m
	return m
}

func (l *linker) determineOffsets() {
	methods := make([]format.Method, len(l.methods))
	var space []format.Implementation

	for n, m := range l.methods {
		slices.SortFunc(m.impls, func(l, r format.Implementation) bool {
			return l.Class < r.Class
		})
		if len(m.impls) == 0 {
			continue
		}
		offset := findOffset(space, m)
		methods[m.id] = format.Method{
			Name:   n,
			Offset: offset,
		}
		space = l.applyOffset(space, m, offset)
	}

	l.program.Implmentations = space
	l.program.Methods = methods
}

func findOffset(space []format.Implementation, m *method) int32 {
next:
	for i := -int(m.impls[0].Class); i < len(space); i++ {
		for _, impl := range m.impls {
			off := int(impl.Class) + i
			if off >= len(space) {
				continue
			}
			if space[off].Class != 0 {
				continue next
			}
		}
		return int32(i)
	}

	return int32(len(space))
}

func (l *linker) applyOffset(space []format.Implementation, m *method, offset int32) []format.Implementation {
	if c := int(offset + int32(m.impls[len(m.impls)-1].Class)); c >= len(space) {
		space = append(space, make([]format.Implementation, 1+c-len(space))...)
	}

	for _, impl := range m.impls {
		space[impl.Class+format.ClassID(offset)] = impl
	}

	return space
}

func writeInt32(into []byte, x int32) {
	for i := 0; i < 4; i++ {
		into[i] = byte(x >> (i * 8))
	}
}
