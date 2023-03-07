package env

import (
	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/runtime/api"
)

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
