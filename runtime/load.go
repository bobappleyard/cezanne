package runtime

import (
	"io"

	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/format/storage"
	"github.com/bobappleyard/cezanne/slices"
)

var externalMethods = map[string]func(*Process){}

func Load(src io.ReaderAt) (*Env, error) {
	var prog format.Program
	err := storage.Read(src, &prog)
	if err != nil {
		return nil, err
	}
	extern := slices.Map(prog.ExternalMethods, func(n string) func(*Process) {
		return externalMethods[n]
	})
	return &Env{
		extern:  extern,
		classes: prog.Classes,
		methods: prog.Bindings,
		code:    prog.Code,
	}, nil
}

func (e *Env) Run() {
	p := &Process{
		env: e,
	}
	p.run()
}

func RegisterExt(name string, fn func(p *Process)) {
	externalMethods[name] = fn
}
