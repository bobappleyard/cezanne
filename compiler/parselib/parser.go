package parselib

import (
	"errors"
	"fmt"
	"io"
)

var ErrUnexpectedToken = errors.New("unexpected token")

type Parser[T any] struct {
	defs map[TokenID]ParserDef[T]
}

type ParserDef[T any] interface {
	Parse(s TokenStream) (T, error)
}

func (p *Parser[T]) Define(tok TokenID, def ParserDef[T]) {
	if p.defs == nil {
		p.defs = map[TokenID]ParserDef[T]{}
	}
	p.defs[tok] = def
}

func (p *Parser[T]) Parse(s TokenStream) (T, error) {
	var zero T

	if !s.Next() {
		if s.Err() == nil {
			return zero, io.ErrUnexpectedEOF
		}
		return zero, s.Err()
	}

	t := s.This()

	if parser, ok := p.defs[t.ID]; ok {
		return parser.Parse(s)
	}

	return zero, fmt.Errorf("%q: %w", s.Text(t), ErrUnexpectedToken)
}

type parserFunc[T any] func(s TokenStream) (T, error)

func (p parserFunc[T]) Parse(s TokenStream) (T, error) {
	return p(s)
}

func expect(s TokenStream, id TokenID) error {
	if !s.Next() {
		if s.Err() == nil {
			return io.ErrUnexpectedEOF
		}
		return s.Err()
	}

	if t := s.This(); t.ID != id {
		return fmt.Errorf("%q: %w", s.Text(t), ErrUnexpectedToken)
	}

	return nil
}

func WithPrefix[T any](d ParserDef[T], prefix TokenID) ParserDef[T] {
	return parserFunc[T](func(s TokenStream) (T, error) {
		var zero T

		if err := expect(s, prefix); err != nil {
			return zero, err
		}

		return d.Parse(s)
	})
}

func WithSuffix[T any](d ParserDef[T], suffix TokenID) ParserDef[T] {
	return parserFunc[T](func(s TokenStream) (T, error) {
		var zero T

		x, err := d.Parse(s)
		if err != nil {
			return zero, err
		}

		if err := expect(s, suffix); err != nil {
			return zero, err
		}

		return x, nil
	})
}
