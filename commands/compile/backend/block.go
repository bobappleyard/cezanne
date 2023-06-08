package backend

import (
	"fmt"

	"github.com/bobappleyard/cezanne/format/assembly"
	"github.com/bobappleyard/cezanne/format/symtab"
)

const baseRegister = 2

type assembler struct {
	syms    *symtab.Symtab
	dest    assembly.Writer
	pending []pendingWork
}

type pendingWork interface {
	doWork(a *assembler)
}

type implementMethod struct {
	class  *assembly.Class
	method method
}

type placeString struct {
	start, end *assembly.Location
	value      string
}

func (w *assembler) writePackage(root method) {
	w.writeBlock(root)
	for len(w.pending) != 0 {
		next := w.pending[0]
		w.pending = w.pending[1:]
		next.doWork(w)
	}
}

func (w *implementMethod) doWork(a *assembler) {
	a.dest.ImplementMethod(w.class, a.dest.Method(w.method.name))
	a.writeBlock(w.method)
}

func (w *placeString) doWork(a *assembler) {
	w.start.Define()
	for _, b := range []byte(w.value) {
		a.dest.WriteByte(int(b))
	}
	w.end.Define()
}

func (w *assembler) writeBlock(src method) {
	for p, s := range src.steps {
		switch s := s.(type) {

		case intStep:
			w.dest.Natural(w.dest.Fixed(s.val))
			w.dest.Store(int(s.into) + baseRegister)

		case stringStep:
			start := w.dest.Location()
			end := w.dest.Location()
			k := w.dest.Location()

			w.dest.Natural(w.dest.Fixed(src.varc + baseRegister))
			w.dest.Store(src.varc + baseRegister)
			w.dest.Natural(k)
			w.dest.Store(src.varc + baseRegister + 1)
			w.dest.Natural(start)
			w.dest.Store(src.varc + baseRegister + 2)
			w.dest.Natural(end)
			w.dest.Store(src.varc + baseRegister + 3)
			w.dest.GlobalLoad(w.dest.Import("runtime"))
			w.dest.Call(w.dest.Method(w.syms.SymbolID("string_constant")), src.varc+baseRegister)
			k.Define()
			w.dest.Store(int(s.into) + baseRegister)
			w.pending = append(w.pending, &placeString{start, end, s.val})

		case localStep:
			w.dest.Load(int(s.from) + baseRegister)
			w.dest.Store(int(s.into) + baseRegister)

		case fieldStep:
			w.dest.Load(int(s.from) + baseRegister)
			w.dest.Field(s.field)
			w.dest.Store(int(s.into) + baseRegister)

		case globalStep:
			w.dest.GlobalLoad(w.dest.Global(s.from))
			w.dest.Store(int(s.into) + baseRegister)

		case importStep:
			w.dest.GlobalLoad(w.dest.Import(s.from))
			w.dest.Store(int(s.into) + baseRegister)

		case createStep:
			for i, f := range s.fields {
				w.dest.Load(int(f) + baseRegister)
				w.dest.Store(src.varc + i + baseRegister)
			}
			c := w.dest.Class(len(s.fields))
			for _, m := range s.methods {
				w.pending = append(w.pending, &implementMethod{c, m})
			}
			w.dest.Create(c, src.varc+baseRegister)
			w.dest.Store(int(s.into) + baseRegister)

		case returnStep:
			w.dest.Load(int(s.val) + baseRegister)
			w.dest.Return()

		case globalStoreStep:
			w.dest.Load(int(s.object) + baseRegister)
			w.dest.GlobalStore(w.dest.Global(s.into))

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
