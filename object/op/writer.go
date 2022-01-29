package op

import "bytes"

type Writer struct {
	buf bytes.Buffer
}

func NewWriter() *Writer {
	return &Writer{}
}

func (w *Writer) Bytes() []byte {
	return w.buf.Bytes()
}

func (w *Writer) Return(data ReturnData) {
	w.writeOp(Op{Opcode: Return})
	w.writeRegister(data.Reg)
}

func (w *Writer) Natural(data NaturalData) {
	n := bytesNeeded(data.Value)
	w.writeOp(Op{Opcode: Natural, Size: n})
	w.writeInt(data.Value, n)
	w.writeRegister(data.Into)
}

func (w *Writer) Local(data LocalData) {
	w.writeOp(Op{Opcode: Local})
	w.writeRegister(data.Source)
	w.writeRegister(data.Into)
}

func (w *Writer) Jump(data JumpData) {
	n := bytesNeeded(data.To)
	w.writeOp(Op{Opcode: Jump, Size: n})
	w.writeInt(data.To, n)
}

func (w *Writer) Branch(data BranchData) {
	n := bytesNeeded(data.To)
	w.writeOp(Op{Opcode: Branch, Size: n})
	w.writeRegister(data.IfNot)
	w.writeInt(data.To, n)
}

func (w *Writer) Call(data CallData) {
	n := bytesNeeded(data.Method)
	w.writeOp(Op{Opcode: Call, Disposition: data.Disposition, Size: n})
	w.writeRegister(data.Into)
	w.writeInt(data.Method, n)
	w.writeRegister(data.Argc)
}

func (w *Writer) CallTail(data CallTailData) {
	n := bytesNeeded(data.Method)
	w.writeOp(Op{Opcode: CallTail, Disposition: data.Disposition, Size: n})
	w.writeRegister(data.Into)
	w.writeInt(data.Method, n)
	w.writeRegister(data.Argc)
}

func (w *Writer) writeOp(op Op) {
	var b byte

	b |= byte(op.Opcode)
	b |= byte(op.Disposition) << 3
	b |= byte(op.Size) << 5

	w.buf.WriteByte(b)
}

func (w *Writer) writeRegister(reg int) {
	w.buf.WriteByte(byte(reg))
}

func (w *Writer) writeInt(x, n int) {
	for n > 0 {
		w.buf.WriteByte(byte(x))
		x >>= 8
		n--
	}
}

func bytesNeeded(x int) int {
	n := 1
	for x > 255 {
		n++
		x >>= 8
	}
	return n
}
