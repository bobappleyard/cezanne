package op

type Reader struct {
	Src []byte
	Pos int
}

func (r *Reader) Op() Op {
	b := r.Src[r.Pos]

	o := (b & 0b00000111)
	d := (b & 0b00011000) >> 3
	s := (b & 0b11100000) >> 5

	return Op{
		Opcode:      Opcode(o),
		Disposition: Disposition(d),
		Size:        int(s),
	}
}

func (r *Reader) Return() ReturnData {
	r.Pos++
	reg := r.readRegister()

	return ReturnData{
		Reg: reg,
	}
}

func (r *Reader) Natural() NaturalData {
	op := r.Op()
	r.Pos++

	value := r.readInt(op.Size)
	into := r.readRegister()

	return NaturalData{
		Value: value,
		Into:  into,
	}
}

func (r *Reader) Local() LocalData {
	r.Pos++

	src := r.readRegister()
	dest := r.readRegister()

	return LocalData{
		Source: src,
		Into:   dest,
	}
}

func (r *Reader) Jump() JumpData {
	op := r.Op()
	r.Pos++

	to := r.readInt(op.Size)

	return JumpData{
		To: to,
	}
}

func (r *Reader) Branch() BranchData {
	op := r.Op()
	r.Pos++

	ifnot := r.readRegister()
	to := r.readInt(op.Size)

	return BranchData{
		IfNot: ifnot,
		To:    to,
	}
}

func (r *Reader) Call() CallData {
	op := r.Op()
	r.Pos++

	recv := r.readRegister()
	methodIdx := r.readInt(op.Size)
	argc := r.readRegister()

	return CallData{
		Method:      methodIdx,
		Into:        recv,
		Argc:        argc,
		Disposition: op.Disposition,
	}
}

func (r *Reader) CallTail() CallTailData {
	op := r.Op()
	r.Pos++

	recv := r.readRegister()
	methodIdx := r.readInt(op.Size)
	argc := r.readRegister()

	return CallTailData{
		Method:      methodIdx,
		Into:        recv,
		Argc:        argc,
		Disposition: op.Disposition,
	}

}

func (r *Reader) readRegister() int {
	b := r.Src[r.Pos]
	r.Pos++
	return int(b)
}

func (r *Reader) readInt(size int) int {
	switch size {
	case 1:
		a := int(r.Src[r.Pos])
		r.Pos++
		return a

	case 2:
		a := int(r.Src[r.Pos])
		a += int(r.Src[r.Pos+1]) << 8
		r.Pos += 2
		return a

	case 3:
		a := int(r.Src[r.Pos])
		a += int(r.Src[r.Pos+1]) << 8
		a += int(r.Src[r.Pos+2]) << 16
		r.Pos += 3
		return a

	case 4:
		a := int(r.Src[r.Pos])
		a += int(r.Src[r.Pos+1]) << 8
		a += int(r.Src[r.Pos+2]) << 16
		a += int(r.Src[r.Pos+3]) << 24
		r.Pos += 4
		return a
	}

	panic("invalid size")
}
