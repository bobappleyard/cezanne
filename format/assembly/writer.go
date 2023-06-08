package assembly

import (
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/format/symtab"
)

type Writer struct {
	syms     *symtab.Symtab
	code     []byte
	rels     []format.Relocation
	methods  []format.Method
	classes  []format.Class
	bindings []format.Implementation
	external []symtab.Symbol
	imports  []string
}

type Value interface {
	write()
}

func (b *Writer) Package() *format.Package {
	return &format.Package{
		ExternalMethods: b.external,
		Imports:         b.imports,
		Classes:         b.classes,
		Methods:         b.methods,
		Implementations: b.bindings,
		Relocations:     b.rels,
		Code:            b.code,
	}
}

func New(syms *symtab.Symtab) *Writer {
	return &Writer{
		syms: syms,
	}
}

func (b *Writer) Load(id int) {
	b.WriteByte(format.LoadOp)
	b.WriteByte(id)
}

func (b *Writer) Store(id int) {
	b.WriteByte(format.StoreOp)
	b.WriteByte(id)
}

func (b *Writer) Natural(value Value) {
	b.WriteByte(format.NaturalOp)
	value.write()
}

func (b *Writer) GlobalLoad(id *Global) {
	b.WriteByte(format.GlobalLoadOp)
	id.write()
}

func (b *Writer) GlobalStore(id *Global) {
	b.WriteByte(format.GlobalStoreOp)
	id.write()
}

func (b *Writer) Create(id *Class, base int) {
	b.WriteByte(format.CreateOp)
	id.write()
	b.WriteByte(base)
}

func (b *Writer) Field(id int) {
	b.WriteByte(format.FieldOp)
	b.writeInt(id)
}

func (b *Writer) Return() {
	b.WriteByte(format.RetOp)
}

func (b *Writer) Call(methodID *Method, base int) {
	b.WriteByte(format.CallOp)
	methodID.write()
	b.WriteByte(base)
}

func (b *Writer) WriteByte(value int) {
	b.code = append(b.code, byte(value))
}

func (b *Writer) writeInt(value int) {
	b.WriteByte(value)
	b.WriteByte(value >> 8)
	b.WriteByte(value >> 16)
	b.WriteByte(value >> 24)
}

type Location struct {
	b    *Writer
	refs []int
	def  int
}

func (l *Location) write() {
	if l.def == 0 {
		l.refs = append(l.refs, len(l.b.rels))
	}
	l.b.rels = append(l.b.rels, format.Relocation{
		Kind: format.CodeRel,
		ID:   int32(l.def),
		Pos:  uint32(len(l.b.code)),
	})
	l.b.writeInt(0)
}

func (l *Location) Define() {
	l.def = len(l.b.code)
	for _, r := range l.refs {
		l.b.rels[r].ID = int32(l.def)
	}
	l.refs = nil
}

func (b *Writer) Location() *Location {
	return &Location{
		b: b,
	}
}

type Method struct {
	b  *Writer
	id format.MethodID
}

func (l *Method) write() {
	l.b.rels = append(l.b.rels, format.Relocation{
		Kind: format.MethodRel,
		ID:   int32(l.id),
		Pos:  uint32(len(l.b.code)),
	})
	l.b.writeInt(0)
}

func (b *Writer) Method(name symtab.Symbol) *Method {
	var id int
	b.methods, id = ensure(b.methods, func(x format.Method) bool {
		return x.Name == name
	}, func() format.Method {
		return format.Method{Name: name, Visibility: format.Public}
	})
	return &Method{
		b:  b,
		id: format.MethodID(id),
	}
}

type Class struct {
	b  *Writer
	id format.ClassID
}

func (l *Class) write() {
	l.b.rels = append(l.b.rels, format.Relocation{
		Kind: format.ClassRel,
		ID:   int32(l.id),
		Pos:  uint32(len(l.b.code)),
	})
	l.b.writeInt(0)
}

func (b *Writer) Class(fieldc int) *Class {
	id := len(b.classes)
	b.classes = append(b.classes, format.Class{Fieldc: uint32(fieldc)})
	return &Class{
		b:  b,
		id: format.ClassID(id),
	}
}

func (c *Class) SetFields(count int) {
	c.b.classes[c.id].Fieldc = uint32(count)
}

type Global struct {
	b    *Writer
	kind format.RelocationKind
	id   int
}

func (l *Global) write() {
	l.b.rels = append(l.b.rels, format.Relocation{
		Kind: l.kind,
		ID:   int32(l.id),
		Pos:  uint32(len(l.b.code)),
	})
	l.b.writeInt(0)
}

func (b *Writer) Global(id int) *Global {
	return &Global{
		b:    b,
		kind: format.GlobalRel,
		id:   id,
	}
}

func (b *Writer) Import(name string) *Global {
	var id int
	b.imports, id = ensure(b.imports, func(x string) bool {
		return x == name
	}, func() string {
		return name
	})
	return &Global{
		b:    b,
		kind: format.ImportRel,
		id:   id,
	}
}

type Fixed struct {
	b     *Writer
	value int
}

func (l *Fixed) write() {
	l.b.writeInt(l.value)
}

func (b *Writer) Fixed(value int) *Fixed {
	return &Fixed{
		b:     b,
		value: value,
	}
}

func (b *Writer) ImplementMethod(class *Class, method *Method) {
	b.bindings = append(b.bindings, format.Implementation{
		Class:      class.id,
		Method:     method.id,
		Kind:       format.StandardBinding,
		EntryPoint: uint32(len(b.code)),
	})
}

func (b *Writer) ImplementExternalMethod(class *Class, method *Method, ext symtab.Symbol) {
	var entryPoint int
	b.external, entryPoint = ensure(b.external, func(x symtab.Symbol) bool {
		return x == ext
	}, func() symtab.Symbol {
		return ext
	})
	b.bindings = append(b.bindings, format.Implementation{
		Class:      class.id,
		Method:     method.id,
		Kind:       format.ExternalBinding,
		EntryPoint: uint32(entryPoint),
	})
}

func ensure[T any](xs []T, test func(x T) bool, cons func() T) ([]T, int) {
	for i, x := range xs {
		if test(x) {
			return xs, i
		}
	}
	res := len(xs)
	xs = append(xs, cons())
	return xs, res
}
