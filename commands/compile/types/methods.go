package types

import (
	"errors"
	"fmt"
	"sort"
)

type Method struct {
	Name string
	In   Type
	Out  Type
	Eff  Type
}

type Shape []Method

var (
	ErrNoMethod = errors.New("no method")
)

func (ms Shape) Get(name string) (Method, error) {
	idx := sort.Search(len(ms), func(i int) bool { return ms[i].Name >= name })
	if idx >= len(ms) || ms[idx].Name != name {
		return Method{}, fmt.Errorf("%s: %w", name, ErrNoMethod)
	}
	return ms[idx], nil
}

func (ms Shape) Copy(seen map[Type]Type) Shape {
	return copySlice(seen, ms)
}

func (ms Shape) Merge(e *Subs, more Shape) (Shape, error) {
	res := make(Shape, len(ms))
	copy(res, ms)
	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })

next:
	for _, m := range more {
		idx := sort.Search(len(res), func(i int) bool { return res[i].Name >= m.Name })
		for ; idx < len(res) && res[idx].Name == m.Name; idx++ {
			n := res[idx]
			if m.In == n.In {
				if err := m.Out.Unify(e, n.Out); err != nil {
					return nil, err
				}
				if err := m.Eff.Unify(e, n.Eff); err != nil {
					return nil, err
				}
				continue next
			}
		}
		res = append(res, Method{})
		copy(res[idx+1:], res[idx:])
		res[idx] = m
	}

	return res, nil
}

func (m Method) Copy(seen map[Type]Type) Method {
	return Method{
		Name: m.Name,
		In:   m.In.Copy(seen),
		Out:  m.Out.Copy(seen),
		Eff:  m.Eff.Copy(seen),
	}
}

func (m Method) Unify(e *Subs, n Method) error {
	if err := n.In.Unify(e, m.In); err != nil {
		return err
	}
	if err := n.Out.Unify(e, m.Out); err != nil {
		return err
	}
	if err := n.Eff.Unify(e, m.Eff); err != nil {
		return err
	}
	return nil
}
