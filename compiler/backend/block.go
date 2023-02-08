package backend

import (
	"fmt"

	"github.com/bobappleyard/cezanne/format/assembly"
)

const baseRegister = 2

type assembler struct {
	dest    assembly.Package
	pending []pendingWork
}

type pendingWork struct {
	class   *assembly.Class
	methods []method
}

func (w *assembler) writePackage(root method) {
	w.writeBlock(root)
	for len(w.pending) != 0 {
		next := w.pending[0]
		w.pending = w.pending[1:]
		for _, m := range next.methods {
			w.dest.ImplementMethod(next.class, w.dest.Method(m.name))
			w.dest.Store(m.argc + baseRegister)
			w.writeBlock(m)
		}
	}
}

func (w *assembler) writeBlock(src method) {
	for p, s := range src.steps {
		switch s := s.(type) {

		case intStep:
			w.dest.Natural(w.dest.Fixed(s.val))
			w.dest.Store(int(s.into) + baseRegister)

		case localStep:
			w.dest.Load(int(s.from) + baseRegister)
			w.dest.Store(int(s.into) + baseRegister)

		case fieldStep:
			w.dest.Load(int(s.from) + baseRegister)
			w.dest.Field(s.field)
			w.dest.Store(int(s.into) + baseRegister)

		case importStep:
			w.dest.GlobalLoad(w.dest.Import(s.path))
			w.dest.Store(int(s.into) + baseRegister)

		case createStep:
			for i, f := range s.fields {
				w.dest.Load(int(f) + baseRegister)
				w.dest.Store(src.varc + i + baseRegister)
			}
			pending := pendingWork{w.dest.Class(len(s.fields)), s.methods}
			w.pending = append(w.pending, pending)
			w.dest.Create(pending.class, src.varc+baseRegister)
			w.dest.Store(int(s.into) + baseRegister)

		case returnStep:
			w.dest.Load(int(s.val) + baseRegister)
			w.dest.Return()

		case storeMainPackageStep:
			w.dest.Load(int(s.object) + baseRegister)
			w.dest.GlobalStore(w.dest.Import("."))

		case callStep:
			if isTailCall(src.steps[p+1:], s.into) {
				w.dest.Load(int(s.object) + baseRegister)
				w.dest.Store(src.varc + baseRegister)

				for i, f := range s.params {
					w.dest.Load(int(f) + baseRegister)
					w.dest.Store(src.varc + i + baseRegister + 1)
				}
				for i := range s.params {
					w.dest.Load(src.varc + i + baseRegister + 1)
					w.dest.Store(i + baseRegister)
				}

				w.dest.Load(src.varc + baseRegister)
				w.dest.Call(w.dest.Method(s.method), 0)
				// we do this to skip the final return instruction
				return

			} else {
				k := w.dest.Location()

				// establish activation frame
				w.dest.Natural(w.dest.Fixed(src.varc + baseRegister))
				w.dest.Store(src.varc + baseRegister)
				w.dest.Natural(k)
				w.dest.Store(src.varc + baseRegister + 1)

				for i, f := range s.params {
					w.dest.Load(int(f) + baseRegister)
					w.dest.Store(src.varc + i + baseRegister*2)
				}
				w.dest.Load(int(s.object) + baseRegister)
				w.dest.Call(w.dest.Method(s.method), src.varc+baseRegister)

				// continuation
				k.Define()
				w.dest.Store(int(s.into) + baseRegister)
			}

		default:
			panic(fmt.Sprintf("unsupported step type: %T", s))

		}
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
