package assembly

import (
	"github.com/bobappleyard/cezanne/format"
)

type Block struct {
	code     []byte
	rels     []format.Relocation
	methods  []format.Method
	classes  []format.Class
	bindings []format.Implementation
	external []string
	imports  []string
}

type Value interface {
	write()
}

func (b *Block) Package() *format.Package {
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

func (b *Block) Load(id int) {
	b.writeByte(format.LoadOp)
	b.writeByte(id)
}

func (b *Block) Store(id int) {
	b.writeByte(format.StoreOp)
	b.writeByte(id)
}

func (b *Block) Natural(value Value) {
	b.writeByte(format.NaturalOp)
	value.write()
}

func (b *Block) GlobalLoad(id *Global) {
	b.writeByte(format.GlobalLoadOp)
	id.write()
}

func (b *Block) GlobalStore(id *Global) {
	b.writeByte(format.GlobalLoadOp)
	id.write()
}

func (b *Block) Create(id *Class, base int) {
	b.writeByte(format.CreateOp)
	id.write()
	b.writeByte(base)
}

func (b *Block) Return() {
	b.writeByte(format.RetOp)
}

func (b *Block) Call(methodID *Method, base int) {
	b.writeByte(format.CallOp)
	methodID.write()
	b.writeByte(base)
}

func (b *Block) writeByte(value int) {
	b.code = append(b.code, byte(value))
}

func (b *Block) writeInt(value int) {
	b.writeByte(value)
	b.writeByte(value >> 8)
	b.writeByte(value >> 16)
	b.writeByte(value >> 24)
}

type Location struct {
	b    *Block
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
		Pos:  int32(len(l.b.code)),
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

func (b *Block) Location() *Location {
	return &Location{
		b: b,
	}
}

type Method struct {
	b  *Block
	id format.MethodID
}

func (l *Method) write() {
	l.b.rels = append(l.b.rels, format.Relocation{
		Kind: format.MethodRel,
		ID:   int32(l.id),
		Pos:  int32(len(l.b.code)),
	})
	l.b.writeInt(0)
}

func (b *Block) Method(name string) *Method {
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
	b  *Block
	id format.ClassID
}

func (l *Class) write() {
	l.b.rels = append(l.b.rels, format.Relocation{
		Kind: format.ClassRel,
		ID:   int32(l.id),
		Pos:  int32(len(l.b.code)),
	})
	l.b.writeInt(0)
}

func (b *Block) Class(name string) *Class {
	var id int
	b.classes, id = ensure(b.classes, func(x format.Class) bool {
		return x.Name == name
	}, func() format.Class {
		return format.Class{Name: name}
	})
	return &Class{
		b:  b,
		id: format.ClassID(id),
	}
}

func (c *Class) SetFields(count int) {
	c.b.classes[c.id].Fieldc = uint32(count)
}

type Global struct {
	b  *Block
	id int
}

func (l *Global) write() {
	l.b.rels = append(l.b.rels, format.Relocation{
		Kind: format.GlobalRel,
		ID:   int32(l.id),
		Pos:  int32(len(l.b.code)),
	})
	l.b.writeInt(0)
}

func (b *Block) Global(name string) *Global {
	var id int
	b.imports, id = ensure(b.imports, func(x string) bool {
		return x == name
	}, func() string {
		return name
	})
	return &Global{
		b:  b,
		id: id,
	}
}

type Fixed struct {
	b     *Block
	value int
}

func (l *Fixed) write() {
	l.b.writeInt(l.value)
}

func (b *Block) Fixed(value int) *Fixed {
	return &Fixed{
		b:     b,
		value: value,
	}
}

func (b *Block) ImplementMethod(class *Class, method *Method) {
	b.bindings = append(b.bindings, format.Implementation{
		Class:      class.id,
		Method:     method.id,
		Kind:       format.StandardBinding,
		EntryPoint: uint32(len(b.code)),
	})
}

func (b *Block) ImplementExternalMethod(class *Class, method *Method, ext string) {
	var entryPoint int
	b.external, entryPoint = ensure(b.external, func(x string) bool {
		return x == ext
	}, func() string {
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
