package linker

import (
	"errors"
	"fmt"

	"github.com/bobappleyard/cezanne/format"
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

type Program struct {
	Included        []Package
	Classes         []Class
	Methods         []format.Method
	Implementations []Implementation
}

type Package struct {
	Name   string
	Path   string
	Offset int
}

type Class struct {
	ID         int
	FieldCount int
}

type Implementation struct {
	Class  int
	Method int
	Symbol string
}

// Given a collection of packages keyed by path, create a program by starting
// with the "main" package and taking the transitive closure of the import
// relation.
func Link(env LinkerEnv, start string) (*Program, error) {
	l := &linker{
		env: env,
		program: Program{
			Classes: []Class{{}},
		},
		imports: map[string]int{},
		methods: map[string]*method{},
	}
	err := l.importPackage(start)
	if err != nil {
		return nil, err
	}
	l.determineOffsets()
	return &l.program, nil
}

type linker struct {
	env     LinkerEnv
	program Program
	imports map[string]int
	methods map[string]*method
}

type method struct {
	id    int
	impls []Implementation
}

func (l *linker) importPackage(path string) error {
	if p, ok := l.imports[path]; ok {
		if p == -1 {
			return ErrCircularImport
		}
		return nil
	}
	l.imports[path] = -1

	p, err := l.env.LoadPackage(path)
	if err != nil {
		return err
	}

	for _, q := range p.Imports {
		err := l.importPackage(q)
		if err != nil {
			return err
		}
	}
	l.imports[path] = len(l.program.Included)
	l.program.Included = append(l.program.Included, Package{
		Name:   p.Name,
		Path:   path,
		Offset: len(l.program.Classes),
	})

	for _, m := range p.Methods {
		l.getMethod(m.Name)
	}

	for i, c := range p.Classes {
		nextClass := len(l.program.Classes)
		l.program.Classes = append(l.program.Classes, Class{
			ID:         nextClass,
			FieldCount: c.FieldCount,
		})
		for _, meth := range c.Methods {
			m := l.getMethod(meth.Name)
			m.impls = append(m.impls, Implementation{
				Class:  nextClass,
				Method: m.id,
				Symbol: fmt.Sprintf("cz_impl_%s_%d_%s", p.Name, i, meth.Name),
			})
		}
	}

	return nil
}

func (l *linker) getMethod(name string) *method {
	if m, ok := l.methods[name]; ok {
		return m
	}
	m := &method{
		id: len(l.methods),
	}
	l.methods[name] = m
	return m
}

func (l *linker) determineOffsets() {
	methods := make([]format.Method, len(l.methods))
	var space []Implementation

	for n, m := range l.methods {
		slices.SortFunc(m.impls, func(l, r Implementation) bool {
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

	l.program.Implementations = space
	l.program.Methods = methods
}

func findOffset(space []Implementation, m *method) int {
next:
	for i := -m.impls[0].Class; i < len(space); i++ {
		for _, impl := range m.impls {
			off := impl.Class + i
			if off >= len(space) {
				continue
			}
			if space[off].Class != 0 {
				continue next
			}
		}
		return i
	}

	return len(space)
}

func (l *linker) applyOffset(space []Implementation, m *method, offset int) []Implementation {
	if c := offset + m.impls[len(m.impls)-1].Class; c >= len(space) {
		space = append(space, make([]Implementation, 1+c-len(space))...)
	}

	for _, impl := range m.impls {
		space[impl.Class+offset] = impl
	}

	return space
}
