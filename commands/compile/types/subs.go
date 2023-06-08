package types

import (
	"sort"
)

type Subs struct {
	equalities table[Type]
}

func (e *Subs) Resolve(t Type) Type {
	for {
		v, ok := t.(*Metavar)
		if !ok {
			return t
		}
		ts := e.equalities.search(v)
		if len(ts) == 0 {
			return t
		}
		t = ts[0]
	}
}

func (e *Subs) VarMeans(v *Metavar, t Type) {
	e.equalities = e.equalities.add(v, t)
}

type table[T any] struct {
	ids  []*Metavar
	vals []T
}

func (t table[T]) search(id *Metavar) []T {
	start, end := t.bounds(id)
	return t.vals[start:end]
}

func (t table[T]) add(id *Metavar, x T) table[T] {
	_, end := t.bounds(id)
	ids := make([]*Metavar, 0, len(t.vals)+1)
	vals := make([]T, 0, len(t.vals)+1)

	ids = append(ids, t.ids[:end]...)
	vals = append(vals, t.vals[:end]...)

	ids = append(ids, id)
	vals = append(vals, x)

	ids = append(ids, t.ids[end:]...)
	vals = append(vals, t.vals[end:]...)

	return table[T]{ids: ids, vals: vals}
}

func (t table[T]) remove(id *Metavar) table[T] {
	start, end := t.bounds(id)

	if start == end {
		return t
	}

	ids := make([]*Metavar, len(t.vals)-(end-start))
	vals := make([]T, len(t.vals)-(end-start))

	ids = append(ids, t.ids[:start]...)
	vals = append(vals, t.vals[:start]...)

	ids = append(ids, t.ids[end:]...)
	vals = append(vals, t.vals[end:]...)

	return table[T]{ids: ids, vals: vals}
}

func (t table[T]) bounds(id *Metavar) (start, end int) {
	start = sort.Search(len(t.vals), func(i int) bool { return t.ids[i].id >= id.id })
	end = sort.Search(len(t.vals), func(i int) bool { return t.ids[i].id > id.id })
	return start, end
}
