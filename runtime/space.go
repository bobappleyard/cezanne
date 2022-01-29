package runtime

import (
	"sort"
)

// Space arranges sets of methods in a manner that balances speed of access with use of space. This
// is achieved using a sparse matrix.
type Space struct {
	methods         map[string]MethodID
	groups          []int
	implementations []implementation
}

type Member struct {
	MethodID       MethodID
	Implementation Object
}

type MethodID int
type ClassID int

type implementation struct {
	classID  ClassID
	methodID MethodID
	group    int
	index    int
}

func (s *Space) Method(name string) MethodID {
	if id, ok := s.methods[name]; ok {
		return id
	}
	id := MethodID(len(s.methods))
	s.methods[name] = id
	return id
}

func (s *Space) Class(class ClassID, methods []MethodID) int {
	implementations := make([]implementation, len(methods))
	for i, m := range methods {
		implementations[i] = implementation{
			classID:  class,
			methodID: m,
			index:    i,
		}
	}
	sort.Slice(implementations, func(i, j int) bool {
		return implementations[i].methodID < implementations[j].methodID
	})
	off := s.determineOffset(implementations)
	s.applyOffset(implementations, off)
	return off
}

func (s *Space) LookupMethod(m MethodID, c ClassID, coff int) (int, error) {
	off := int(m) + coff
	if off < 0 || off >= len(s.implementations) {
		return 0, ErrUnknownMember
	}
	p := s.implementations[off]
	if p.methodID != m || p.classID != c {
		return 0, ErrUnknownMember
	}
	return p.index, nil
}

func (s *Space) determineOffset(implementations []implementation) int {
	off := int(-implementations[0].methodID)
	for {
		delta := 0
		for _, m := range implementations {
			p := int(m.methodID) + off
			d := s.nextFreePos(p) - p
			if d > delta {
				delta = d
			}
		}
		if delta == 0 {
			break
		}
		off += delta
	}
	return off
}

func (s *Space) applyOffset(implementations []implementation, off int) {
	n := 1 + off + int(implementations[len(implementations)-1].methodID)
	if n > len(s.implementations) {
		s.implementations = append(s.implementations, make([]implementation, n-len(s.implementations))...)
	}

	// need to go backwards because filling in a location could merge two groups
	for i := len(implementations) - 1; i >= 0; i-- {
		m := implementations[i]
		p := off + int(m.methodID)
		s.implementations[p] = m
		s.fillPos(p)
	}
}

func (s *Space) nextFreePos(p int) int {
	if p >= len(s.implementations) {
		return p
	}
	group := s.implementations[p].group
	if group == 0 {
		return p
	}
	return s.groups[group-1]
}

func (s *Space) fillPos(p int) {
	var before, after int

	if p > 0 {
		before = s.implementations[p-1].group
	}
	if p < len(s.implementations)-1 {
		after = s.implementations[p+1].group
	}

	switch {
	case before == 0 && after == 0:
		s.groups = append(s.groups, p+1)
		s.implementations[p].group = len(s.groups)

	case before == 0:
		s.implementations[p].group = after

	case after == 0:
		s.implementations[p].group = before
		s.groups[before-1] = p + 1

	default:
		s.implementations[p].group = before
		s.groups[before-1] = s.groups[after-1]
	}
}
