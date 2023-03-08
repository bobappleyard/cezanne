package env

import (
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/runtime/api"
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
