package backend

import (
	"fmt"
	"sort"

	"github.com/bobappleyard/cezanne/slices"
)

// reuse registers when safe to do so
func reuseVariables(m *method, isMethod bool) {
	vs := varSpace{argc: m.argc}
	if isMethod {
		vs.argc++
	}
	determineBounds(&vs, m)
	updateBounds(&vs, m)

	m.varc = int(vs.maxVar) + 1
}

type varBound struct {
	v, new variable
	end    int
}

type varSpace struct {
	argc    int
	maxVar  variable
	current int        // current instruction
	vars    []int      // mapping from variable ID to most recent creation
	bounds  []varBound // mapping from instruction to binding info
}

func (s *varSpace) restart() {
	s.current = -1
}

func (s *varSpace) next() {
	if s.current >= 0 {
		s.vars[s.bounds[s.current].v] = s.current
	}
	s.current++
	if s.current >= len(s.bounds) {
		s.bounds = append(s.bounds, make([]varBound, 1+s.current-len(s.bounds))...)
	}
}

func (s *varSpace) createBound(v variable) {
	if int(v) >= len(s.vars) {
		s.vars = append(s.vars, make([]int, 1+int(v)-len(s.vars))...)
	}
	s.bounds[s.current].v = v
}

func (s *varSpace) consumeBound(v variable) {
	if int(v) < s.argc {
		return
	}
	s.bounds[s.vars[v]].end = s.current
}

func (s *varSpace) consumeBounds(vs []variable) {
	for _, v := range vs {
		s.consumeBound(v)
	}
}

func (s *varSpace) liveBounds() []variable {
	var found []variable
	for _, b := range s.bounds[:s.current] {
		if b.end > s.current {
			found = append(found, b.v)
		}
	}
	return found
}

func (s *varSpace) remapVariable() variable {
	live := s.liveBounds()
	sort.Slice(live, func(i, j int) bool { return live[i] < live[j] })

	if len(live) == 0 {
		return s.establishRemapping(variable(s.argc))
	}

	last := variable(s.argc)
	for _, v := range live {
		if v <= last+1 {
			// no space for a new variable
			last = v
			continue
		}

		return s.establishRemapping(last + 1)

	}

	return s.establishRemapping(live[len(live)-1] + 1)
}

func (s *varSpace) establishRemapping(to variable) variable {
	if to > s.maxVar {
		s.maxVar = to
	}
	s.bounds[s.current].new = to
	return to
}

func (s *varSpace) getVarMapping(v variable) variable {
	if int(v) < s.argc {
		return v
	}
	to := s.bounds[s.vars[v]].new
	return to
}

func determineBounds(vs *varSpace, m *method) {

	vs.restart()

	for _, s := range m.steps {
		vs.next()

		switch s := s.(type) {
		case intStep:
			vs.createBound(s.into)

		case localStep:
			vs.consumeBound(s.from)
			vs.createBound(s.into)

		case fieldStep:
			vs.consumeBound(s.from)
			vs.createBound(s.into)

		case createStep:
			vs.consumeBounds(s.fields)
			vs.createBound(s.into)

		case returnStep:
			vs.consumeBound(s.val)

		case callMethodStep:
			vs.consumeBound(s.object)
			vs.consumeBounds(s.params)
			vs.createBound(s.into)

		case callFunctionStep:
			vs.consumeBounds(s.params)
			vs.createBound(s.into)

		default:
			panic(fmt.Sprintf("unknown step type %#v", s))
		}
	}
}

func updateBounds(vs *varSpace, m *method) {
	vs.restart()

	for i, s := range m.steps {
		vs.next()

		switch s := s.(type) {
		case intStep:
			s.into = vs.remapVariable()

			m.steps[i] = s

		case localStep:
			s.from = vs.getVarMapping(s.from)
			s.into = vs.remapVariable()

			m.steps[i] = s

		case fieldStep:
			s.from = vs.getVarMapping(s.from)
			s.into = vs.remapVariable()

			m.steps[i] = s

		case createStep:
			s.fields = slices.Map(s.fields, vs.getVarMapping)
			s.into = vs.remapVariable()

			m.steps[i] = s

		case returnStep:
			s.val = vs.getVarMapping(s.val)

			m.steps[i] = s

		case callMethodStep:
			s.object = vs.getVarMapping(s.object)
			s.params = slices.Map(s.params, vs.getVarMapping)
			s.into = vs.remapVariable()

			m.steps[i] = s

		case callFunctionStep:
			s.params = slices.Map(s.params, vs.getVarMapping)
			s.into = vs.remapVariable()

			m.steps[i] = s

		default:
			panic(fmt.Sprintf("unknown step type %#v", s))
		}
	}

}
