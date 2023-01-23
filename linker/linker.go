package linker

import (
	"github.com/bobappleyard/cezanne/format"
	"golang.org/x/exp/slices"
)

type Linker struct {
	program format.Program
	methods map[string]*method
}

type method struct {
	id    format.MethodID
	impls []format.Implementation
}

func (l *Linker) Complete() *format.Program {
	l.determineOffsets()
	return &l.program
}

func (l *Linker) AddPackage(p *format.Package) error {
	if len(l.program.Classes) == 0 {
		l.methods = map[string]*method{}
		l.program.Classes = make([]format.Class, 1)
	}
	l.processRelocations(p)
	l.processBindings(p)
	l.program.ExternalMethods = append(l.program.ExternalMethods, p.ExternalMethods...)
	l.program.Classes = append(l.program.Classes, p.Classes...)
	l.program.Code = append(l.program.Code, p.Code...)
	return nil
}

func (l *Linker) processRelocations(p *format.Package) {
	var glob int32
	for _, rel := range p.Relocations {
		switch rel.Kind {
		case format.ClassRel:
			rel.ID += int32(len(l.program.Classes))

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
		for i := 0; i < 4; i++ {
			p.Code[i+int(rel.Pos)] = byte(rel.ID >> (i * 8))
		}
	}
	l.program.GlobalCount += glob
}

func (l *Linker) processBindings(p *format.Package) {
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

func (l *Linker) method(name string) *method {
	if m, ok := l.methods[name]; ok {
		return m
	}
	m := &method{id: format.MethodID(len(l.methods))}
	l.methods[name] = m
	return m
}

func (l *Linker) determineOffsets() {
	offsets := make([]int32, len(l.methods))
	var space []format.Implementation

	for _, m := range l.methods {
		slices.SortFunc(m.impls, func(l, r format.Implementation) bool {
			return l.Class < r.Class
		})
		offset := findOffset(space, m)
		offsets[m.id] = offset
		space = l.applyOffset(space, m, offset)
	}

	l.program.Implmentations = space
	l.program.MethodOffsets = offsets
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

func (l *Linker) applyOffset(space []format.Implementation, m *method, offset int32) []format.Implementation {
	if c := int(offset + int32(m.impls[len(m.impls)-1].Class)); c >= len(space) {
		space = append(space, make([]format.Implementation, 1+c)...)
	}

	for _, impl := range m.impls {
		space[impl.Class+format.ClassID(offset)] = impl
	}

	return space
}
