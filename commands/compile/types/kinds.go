package types

import (
	"errors"
	"sync/atomic"

	"github.com/bobappleyard/cezanne/util/slices"
)

type Type interface {
	Apply(e *Subs) Type
	Copy(seen map[Type]Type) Type

	Unify(e *Subs, u Type) error
	Supports(e *Subs, s Shape) error
}

var (
	ErrWrongKind     = errors.New("wrong kind")
	ErrWrongCons     = errors.New("wrong constructor")
	ErrWrongArgCount = errors.New("wrong number of type args")
	ErrWrongMethods  = errors.New("wrong methods")
)

// Metavar enables polymorphism. It represents a currently-unknown type.
type Metavar struct {
	id         uint64
	constraint Shape
}

var nextVar uint64

func NewVar() *Metavar {
	return &Metavar{
		id: atomic.AddUint64(&nextVar, 1),
	}
}

func (t *Metavar) Apply(e *Subs) Type {
	return e.Resolve(t)
}

func (t *Metavar) Unify(e *Subs, u Type) error {
	if tr := e.Resolve(t); t != tr {
		return tr.Unify(e, u)
	}
	u = e.Resolve(u)

	if t == u {
		return nil
	}

	e.VarMeans(t, u)
	if err := u.Supports(e, t.constraint); err != nil {
		return err
	}

	return nil
}

func (t *Metavar) Supports(e *Subs, s Shape) error {
	if len(s) == 0 {
		return nil
	}
	rv := e.Resolve(t)
	if t != rv {
		return rv.Supports(e, s)
	}
	ss, err := t.constraint.Merge(e, s)
	if err != nil {
		return err
	}
	u := NewVar()
	u.constraint = ss
	e.VarMeans(t, u)
	return nil
}

func (t *Metavar) Copy(seen map[Type]Type) Type {
	if t := seen[t]; t != nil {
		return t
	}
	u := NewVar()
	u.constraint = t.constraint.Copy(seen)

	return u
}

type Constructor struct {
	Name    string
	Args    []Type
	Methods Shape
}

func (c *Constructor) WithArgs(e *Subs, args []Type) (Type, error) {
	if len(args) != len(c.Args) {
		return nil, ErrWrongArgCount
	}
	seen := map[Type]Type{}
	for i, a := range args {
		if err := a.Unify(e, c.Args[i].Copy(seen)); err != nil {
			return nil, err
		}
	}
	return &Named{Cons: c, Args: args}, nil
}

// Named enables nominal typing. It represents an instance of a class.
type Named struct {
	Cons *Constructor
	Args []Type
}

func (t *Named) Apply(e *Subs) Type {
	args := slices.Map(t.Args, func(a Type) Type { return a.Apply(e) })
	for i, a := range args {
		if a != t.Args[i] {
			return &Named{
				Cons: t.Cons,
				Args: args,
			}
		}
	}
	return t
}

func (t *Named) Unify(e *Subs, u Type) error {
	if t == u {
		return nil
	}
	switch u := u.(type) {
	case *Metavar:
		return u.Unify(e, t)

	case *Named:
		if u.Cons != t.Cons {
			return ErrWrongCons
		}
		for i, t := range t.Args {
			u := u.Args[i]

			if err := t.Unify(e, u); err != nil {
				return err
			}

		}
		return nil

	default:
		return ErrWrongKind
	}
}

func (t *Named) Supports(e *Subs, ms Shape) error {
	for _, m := range ms {
		n, err := t.Cons.Methods.Get(m.Name)
		if err != nil {
			return err
		}
		seen := createSeen(t.Args)
		if err := m.Unify(e, n.Copy(seen)); err != nil {
			return err
		}
	}
	return nil
}

func (t *Named) Copy(seen map[Type]Type) Type {
	args := copySlice(seen, t.Args)
	for i, a := range args {
		if a != t.Args[i] {
			return &Named{Cons: t.Cons, Args: args}
		}
	}
	return t
}

type Anonymous struct {
	Methods Shape
	Scope   []Type
}

func (t *Anonymous) Unify(e *Subs, u Type) error {
	switch u := u.(type) {
	case *Metavar:
		return u.Unify(e, t)

	case *Anonymous:
		if len(t.Methods) != len(u.Methods) {
			return ErrWrongMethods
		}

		for i, m := range t.Methods {
			n := u.Methods[i]

			if m.Name != n.Name {
				return ErrWrongMethods
			}

			seenT := createSeen(t.Scope)
			seenU := createSeen(u.Scope)
			if err := m.Copy(seenT).Unify(e, n.Copy(seenU)); err != nil {
				return err
			}
		}

		return nil

	default:
		return ErrWrongKind
	}
}

func (t *Anonymous) Supports(e *Subs, s Shape) error {
	for _, m := range s {
		n, err := t.Methods.Get(m.Name)
		if err != nil {
			return err
		}
		seen := createSeen(t.Scope)

		if err := m.Unify(e, n.Copy(seen)); err != nil {
			return err
		}
	}

	return nil
}

func (t *Anonymous) Copy(seen map[Type]Type) Type {
	return &Anonymous{
		Scope:   copySlice(seen, t.Scope),
		Methods: copySlice(seen, t.Methods),
	}
}

func (t *Anonymous) Apply(e *Subs) Type {
	panic("unimplemented")
}

func createSeen(args []Type) map[Type]Type {
	seen := map[Type]Type{}
	for _, t := range args {
		seen[t] = t
	}
	return seen
}

type Copier[T any] interface {
	Copy(seen map[Type]Type) T
}

func copySlice[T Copier[T]](seen map[Type]Type, ts []T) []T {
	return slices.Map(ts, func(t T) T { return t.Copy(seen) })
}
