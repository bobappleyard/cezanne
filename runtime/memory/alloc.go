package memory

import (
	"math"

	"github.com/bobappleyard/cezanne/format"
	"github.com/bobappleyard/cezanne/runtime/api"
)

type Arena struct {
	env         Env
	front, back []api.Object
	allocated   api.Ref
}

type Collection struct {
	arena  *Arena
	copied api.Ref
}

type Env interface {
	FieldCount(class format.ClassID) int
	MarkRoots(c *Collection)
}

func NewArena(env Env, size int) *Arena {
	return &Arena{
		env:   env,
		front: make([]api.Object, size),
		back:  make([]api.Object, size),
	}
}

func (a *Arena) Alloc(class format.ClassID, fields []api.Object) api.Object {
	res, ok := a.tryAlloc(class, fields)
	if ok {
		return res
	}
	a.collect()
	res, ok = a.tryAlloc(class, fields)
	if ok {
		return res
	}
	panic("out of memory")
}

func (a *Arena) tryAlloc(class format.ClassID, fields []api.Object) (api.Object, bool) {
	fieldCount := api.Ref(a.env.FieldCount(class))
	if fieldCount == 0 {
		return api.Object{
			Class: class,
		}, true
	}
	if a.allocated+fieldCount > api.Ref(len(a.front)) {
		return api.Object{}, false
	}
	res := api.Object{
		Class: class,
		Data:  a.allocated,
	}
	copy(a.front[a.allocated:a.allocated+fieldCount], fields)
	a.allocated += fieldCount
	return res, true
}

func (a *Arena) Get(object api.Object, field int) api.Object {
	return a.front[object.Data+api.Ref(field)]
}

func (a *Arena) Set(object api.Object, field int, value api.Object) {
	a.front[object.Data+api.Ref(field)] = value
}

func (a *Arena) collect() {
	c := &Collection{arena: a}
	a.front, a.back = a.back, a.front
	a.allocated = 0
	a.env.MarkRoots(c)
	c.collect()
}

func (c *Collection) collect() {
	for c.copied < c.arena.allocated {
		c.arena.front[c.copied] = c.Copy(c.arena.front[c.copied])
		c.copied++
	}
}

const reloc = format.ClassID(math.MaxInt32)

func (c *Collection) Copy(old api.Object) api.Object {
	fields := c.arena.env.FieldCount(old.Class)
	if fields == 0 {
		return old
	}
	if c.arena.back[old.Data].Class == reloc {
		return api.Object{
			Class: old.Class,
			Data:  c.arena.back[old.Data].Data,
		}
	}
	new, ok := c.arena.tryAlloc(old.Class, c.arena.back[old.Data:old.Data+api.Ref(fields)])
	if !ok {
		panic("out of memory")
	}
	c.arena.back[old.Data] = api.Object{
		Class: reloc,
		Data:  new.Data,
	}
	return new
}
