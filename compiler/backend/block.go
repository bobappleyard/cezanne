package backend

import (
	"fmt"

	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/slices"
)

const baseRegister variable = 0

type assembler struct {
	dest    assemblyWriter
	pending []pendingWork
}

type pendingWork struct {
	class   int
	methods []method
}

func (w *assembler) writeFunction(s scope, f ast.Method) {
	w.dest.ImplementFunction(f.Name, func() {
		w.writeBlock(interpretFunction(s, f))
	})
}

func (w *assembler) processPending() {
	for len(w.pending) != 0 {
		next := w.pending[0]
		w.pending = w.pending[1:]
		for _, m := range next.methods {
			w.dest.ImplementMethod(next.class, m.name, func() {
				w.dest.Store(baseRegister.offset(m.argc))
				w.writeBlock(m)
			})
		}
	}
}

func (w *assembler) writePackageInit() {
	fmt.Fprintf(&w.dest.code, "extern void cz_impl_%s() {\n", w.dest.meta.Name)
	fmt.Fprintln(&w.dest.code, "}")
	fmt.Fprintln(&w.dest.code)

}

func (w *assembler) writeBlock(src method) {
	for p, s := range src.steps {
		switch s := s.(type) {

		case intStep:
			w.dest.Natural(s.val)
			w.dest.Store(s.into + baseRegister)

		case localStep:
			w.dest.Load(s.from + baseRegister)
			w.dest.Store(s.into + baseRegister)

		case fieldStep:
			w.dest.Load(s.from + baseRegister)
			w.dest.Field(s.field)
			w.dest.Store(s.into + baseRegister)

		case createStep:
			for i, f := range s.fields {
				w.dest.Load(f + baseRegister)
				w.dest.Store(baseRegister.offset(src.varc + i))
			}
			classID := len(w.dest.meta.Classes)
			w.dest.meta.Classes = append(w.dest.meta.Classes, format.Class{
				FieldCount: len(s.fields),
				Methods: slices.Map(s.methods, func(x method) format.Method {
					return format.Method{Name: x.name}
				}),
			})
			pending := pendingWork{classID, s.methods}
			w.pending = append(w.pending, pending)
			w.dest.Create(classID, baseRegister.offset(src.varc))
			w.dest.Store(s.into + baseRegister)

		case returnStep:
			w.dest.Load(s.val + baseRegister)
			w.dest.Return()

		case callMethodStep:
			if isTailCall(src.steps[p+1:], s.into) {
				w.dest.Load(s.object + baseRegister)
				w.dest.Store(baseRegister.offset(src.varc))
				w.compileTailArgs(src, s.params)
				w.dest.Load(baseRegister.offset(src.varc))
				w.dest.FunctionCallTail(s.method)
				// we do this to skip the final return instruction
				return

			} else {
				for i, f := range s.params {
					w.dest.Load(f + baseRegister)
					w.dest.Store(baseRegister.offset(src.varc + i))
				}
				w.dest.Load(s.object + baseRegister)
				w.dest.FunctionCall(s.method, baseRegister.offset(src.varc))
				w.dest.Store(s.into + baseRegister)
			}

		case callFunctionStep:
			if isTailCall(src.steps[p+1:], s.into) {
				w.compileTailArgs(src, s.params)
				w.dest.FunctionCallTail(s.method)
				// we do this to skip the final return instruction
				return

			} else {
				for i, f := range s.params {
					w.dest.Load(f + baseRegister)
					w.dest.Store(baseRegister.offset(src.varc + i))
				}
				w.dest.FunctionCall(s.method, baseRegister.offset(src.varc))
				w.dest.Store(s.into + baseRegister)
			}

		default:
			panic(fmt.Sprintf("unsupported step type: %T", s))

		}
	}
}

func (w *assembler) compileTailArgs(src method, params []variable) {
	for i, f := range params {
		w.dest.Load(f + baseRegister)
		w.dest.Store(baseRegister.offset(src.varc+i) + 1)
	}
	for i := range params {
		w.dest.Load(baseRegister.offset(src.varc+i) + 1)
		w.dest.Store(baseRegister.offset(i))
	}
}

func isTailCall(steps []step, id variable) bool {
	if len(steps) != 1 {
		return false
	}
	s, ok := steps[0].(returnStep)
	if !ok {
		return false
	}
	return s.val == id
}
