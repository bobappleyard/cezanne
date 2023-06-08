package env

import (
	"unicode/utf8"

	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/runtime/api"
	"github.com/bobappleyard/cezanne/util/slices"
)

func (p *Process) Create(class format.ClassID, fields ...api.Object) api.Object {
	return p.memory.Alloc(class, fields)
}

func (p *Process) Field(x api.Object, id int) api.Object {
	return p.memory.Get(x, id)
}

func (p *Process) Int(x int) api.Object {
	return api.Object{
		Class: p.kinds[format.IntKind],
		Data:  api.Ref(x),
	}
}

func (p *Process) AsInt(x api.Object) int {
	return int(x.Data)
}

func (p *Process) Bool(x bool) api.Object {
	if x {
		return api.Object{Class: p.kinds[format.TrueKind]}
	}
	return api.Object{Class: p.kinds[format.FalseKind]}
}

func (p *Process) AsBool(x api.Object) bool {
	return x.Class != p.kinds[format.FalseKind]
}

func (p *Process) Array(items []api.Object) api.Object {
	return p.memory.Alloc(format.ClassID(-len(items)-1), items)
}

func (p *Process) AsArray(x api.Object) []api.Object {
	n := -int(x.Class) - 1
	items := make([]api.Object, n)
	for i := range items {
		items[i] = p.Field(x, i)
	}
	return items
}

func (p *Process) String(s string) api.Object {
	var runes []api.Object

	pos := 0
	for pos < len(s) {
		c, size := utf8.DecodeRuneInString(s[pos:])
		runes = append(runes, p.Int(int(c)))
		pos += size
	}

	return p.Create(p.kinds[format.StringKind], p.Array(runes))
}

func (p *Process) AsString(x api.Object) string {
	return string(slices.Map(p.AsArray(p.Field(x, 0)), func(x api.Object) rune {
		return rune(p.AsInt(x))
	}))
}
