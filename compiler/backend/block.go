package backend

import (
	"fmt"

	"github.com/bobappleyard/cezanne/compiler/ast"
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/slices"
)

type assembler struct {
	dest    assemblyWriter
	pending []pendingWork
}

type pendingWork struct {
	class   int
	methods []method
}

func (w *assembler) writeFunction(s scope, f ast.Method) {
	fn := interpretFunction(s, f)
	w.dest.ImplementFunction(f.Name, fn.argc, fn.usedSpace(), func() {
		reuseVariables(&fn, false)
		w.writeBlock(fn)
	})
}

func (w *assembler) processPending() {
	for len(w.pending) != 0 {
		next := w.pending[0]
		w.pending = w.pending[1:]
		for _, m := range next.methods {
			w.dest.ImplementMethod(next.class, m.name, m.argc, m.usedSpace(), func() {
				reuseVariables(&m, true)
				w.dest.Store(variable(m.argc))
				w.writeBlock(m)
			})
		}
	}
}

func (w *assembler) writePackageInit() {
	w.dest.ImplementInit(0, func() {
		// code to initialise globals goes here
		w.dest.Return()
	})
}

func (w *assembler) writeBlock(src method) {

	for p, s := range src.steps {
		switch s := s.(type) {

		case intStep:
			w.dest.Natural(s.val)
			w.dest.Store(s.into)

		case localStep:
			w.dest.Load(s.from)
			w.dest.Store(s.into)

		case fieldStep:
			w.dest.Load(s.from)
			w.dest.Field(s.field)
			w.dest.Store(s.into)

		case createStep:
			for i, f := range s.fields {
				w.dest.Load(f)
				w.dest.Store(variable(src.varc + i))
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
			w.dest.Create(classID, src.varc)
			w.dest.Store(s.into)

		case returnStep:
			w.dest.Load(s.val)
			w.dest.Return()

		case callMethodStep:
			w.dest.addMethod(s.method)
			method := fmt.Sprintf("cz_m_%s", s.method)

			if isTailCall(src.steps[p+1:], s.into) {
				w.dest.Load(s.object)
				w.dest.Store(variable(src.varc))
				w.compileTailArgs(src, s.params)
				w.dest.Load(variable(src.varc))
				w.dest.Call(method, 0)
				// we do this to skip the final return instruction
				return

			} else {
				for i, f := range s.params {
					w.dest.Load(f)
					w.dest.Store(variable(src.varc + i))
				}
				w.dest.Load(s.object)
				w.dest.Call(method, src.varc)
				w.dest.Store(s.into)
			}

		case callFunctionStep:
			if isTailCall(src.steps[p+1:], s.into) {
				w.compileTailArgs(src, s.params)
				w.dest.Call(s.method, 0)
				// we do this to skip the final return instruction
				return

			} else {
				for i, f := range s.params {
					w.dest.Load(f)
					w.dest.Store(variable(src.varc + i))
				}
				w.dest.Call(s.method, src.varc)
				w.dest.Store(s.into)
			}

		default:
			panic(fmt.Sprintf("unsupported step type: %T", s))

		}
	}
}

func (w *assembler) compileTailArgs(src method, params []variable) {
	for i, f := range params {
		w.dest.Load(f)
		w.dest.Store(variable(src.varc+i) + 1)
	}
	for i := range params {
		w.dest.Load(variable(src.varc+i) + 1)
		w.dest.Store(variable(i))
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
