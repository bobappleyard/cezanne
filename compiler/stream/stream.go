package stream

type Iter[T any] interface {
	This() T
	Next() bool
	Err() error
}

type LL1Stream[T any] struct {
	src     Iter[T]
	useLast bool
}

func LL1[T any](s Iter[T]) *LL1Stream[T] {
	return &LL1Stream[T]{src: s}
}

// Err implements Stream
func (s *LL1Stream[T]) Err() error {
	return s.src.Err()
}

// Next implements Stream
func (s *LL1Stream[T]) Next() bool {
	if s.useLast {
		s.useLast = false
		return true
	}
	return s.src.Next()
}

// This implements Stream
func (s *LL1Stream[T]) This() T {
	return s.src.This()
}

func (s *LL1Stream[T]) Back() {
	s.useLast = true
}

type sliceStream[T any] struct {
	items []T
	pos   int
}

func FromSlice[T any](xs []T) Iter[T] {
	return &sliceStream[T]{items: xs, pos: -1}
}

func Of[T any](xs ...T) Iter[T] {
	return FromSlice(xs)
}

// Err implements Stream
func (s *sliceStream[T]) Err() error {
	return nil
}

// Next implements Stream
func (s *sliceStream[T]) Next() bool {
	s.pos++
	return s.pos < len(s.items)
}

// This implements Stream
func (s *sliceStream[T]) This() T {
	return s.items[s.pos]
}

type mappedStream[T, U any] struct {
	source  Iter[T]
	current Iter[U]
	mapping func(T) Iter[U]
}

func FlatMap[T, U any](src Iter[T], proc func(T) Iter[U]) Iter[U] {
	return &mappedStream[T, U]{
		source:  src,
		current: FromSlice[U](nil),
		mapping: proc,
	}
}

func Map[T, U any](src Iter[T], proc func(T) U) Iter[U] {
	return FlatMap(src, func(x T) Iter[U] { return Of(proc(x)) })
}

func Filter[T any](src Iter[T], test func(T) bool) Iter[T] {
	return FlatMap(src, func(x T) Iter[T] {
		if test(x) {
			return Of(x)
		}
		return Of[T]()
	})
}

// Err implements Stream
func (s *mappedStream[T, U]) Err() error {
	if s.current.Err() != nil {
		return s.current.Err()
	}
	if s.source.Err() != nil {
		return s.source.Err()
	}
	return nil
}

// Next implements Stream
func (s *mappedStream[T, U]) Next() bool {
start:
	if s.current.Next() {
		return true
	}
	if s.current.Err() != nil {
		return false
	}
	if !s.source.Next() {
		return false
	}
	s.current = s.mapping(s.source.This())
	goto start
}

// This implements Stream
func (s *mappedStream[T, U]) This() U {
	return s.current.This()
}
