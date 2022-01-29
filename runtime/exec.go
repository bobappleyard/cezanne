package runtime

import (
	"github.com/bobappleyard/cezanne/object/op"
)

type block struct {
	unit       *Unit
	argc, varc int
	entryPoint int
}

func (b *block) CallMethod(p *Process) Action {
	if p.MethodID() != callMethod {
		return p.Error(ErrUnknownMethod)
	}
	if p.Argc() != b.argc+1 {
		return p.Error(ErrWrongArgCount)
	}

	// Doing this makes the compiler allocate the method's variables on the stack
	switch {
	case b.varc < 8:
		return b.exec8(p)
	case b.varc < 16:
		return b.exec16(p)
	case b.varc < 32:
		return b.exec32(p)
	case b.varc < 64:
		return b.exec64(p)
	case b.varc < 128:
		return b.exec128(p)
	default:
		return b.exec256(p)
	}
}

func (b *block) exec8(p *Process) Action   { return b.execInstructions(p, make([]Object, 8)) }
func (b *block) exec16(p *Process) Action  { return b.execInstructions(p, make([]Object, 16)) }
func (b *block) exec32(p *Process) Action  { return b.execInstructions(p, make([]Object, 32)) }
func (b *block) exec64(p *Process) Action  { return b.execInstructions(p, make([]Object, 64)) }
func (b *block) exec128(p *Process) Action { return b.execInstructions(p, make([]Object, 128)) }
func (b *block) exec256(p *Process) Action { return b.execInstructions(p, make([]Object, 256)) }

func (b *block) execInstructions(p *Process, vars []Object) Action {
	vars[0] = p.Wrap(b.unit)

	for i := 0; i < p.Argc(); i++ {
		vars[i+1] = p.Arg(i)
	}

	r := op.Reader{
		Src: b.unit.Code,
		Pos: b.entryPoint,
	}

	for {
		opcode := r.Op()

		switch opcode.Opcode {
		case op.Return:
			data := r.Return()

			return p.Return(vars[data.Reg])

		case op.Natural:
			data := r.Natural()

			vars[data.Into] = FromInt(data.Value)

		case op.Local:
			data := r.Local()

			vars[data.Into] = vars[data.Source]

		case op.Jump:
			data := r.Jump()

			r.Pos = data.To

		case op.Branch:
			data := r.Branch()

			if !Bool(vars[data.IfNot]) {
				r.Pos = data.To
			}

		case op.Call:
			data := r.Call()

			args := vars[data.Into+1 : data.Into+1+data.Argc]
			res, err := p.Call(vars[data.Into], b.unit.Methods[data.Method], args...)
			if err != nil {
				return p.Error(err)
			}

			vars[data.Into] = res

		case op.CallTail:
			data := r.Call()

			args := vars[data.Into+1 : data.Into+1+data.Argc]
			return p.CallTail(vars[data.Into], b.unit.Methods[data.Method], args...)

		default:
			panic("invalid bytecode")
		}
	}
}
