package text

import (
	"unicode/utf8"
)

type LexerState int

// Lexer is a simple Thompson-style NFA.
//
// It maintains a description of a state machine where movement between states is driven by reading
// an input text.
type Lexer[T any] struct {
	closeTransitions []closeTransition
	moveTransitions  []moveTransition
	finalStates      []finalState[T]
	maxState         LexerState
}

type TokenSpec[T any] func(l *Lexer[T]) error

func NewLexer[T any](tokens ...TokenSpec[T]) (*Lexer[T], error) {
	l := new(Lexer[T])
	for _, s := range tokens {
		if err := s(l); err != nil {
			return nil, err
		}
	}
	return l, nil
}

type closeTransition struct {
	Given, Then LexerState
}

type moveTransition struct {
	Given, Then LexerState
	Min, Max    rune
}

type finalState[T any] struct {
	Given LexerState
	Then  func(start int, text string) T
}

type Stream[T any] struct {
	prog       *Lexer[T]
	src        []byte
	srcPos     int
	this, next []bool
	tok        T
	err        error
}

func (p *Lexer[T]) State() LexerState {
	p.maxState++
	return p.maxState
}

func (p *Lexer[T]) Rune(given, then LexerState, r rune) {
	p.Range(given, then, r, r)
}

func (p *Lexer[T]) Range(given, then LexerState, min, max rune) {
	p.moveTransitions = append(p.moveTransitions, moveTransition{
		Given: given,
		Then:  then,
		Min:   min,
		Max:   max,
	})
}

func (p *Lexer[T]) Empty(given, then LexerState) {
	var pending []closeTransition
	for _, t := range p.closeTransitions {
		// avoid adding duplicates
		if t.Given == given && t.Then == then {
			return
		}
		// ensure transitive property is maintained
		if t.Given == then {
			pending = append(pending, closeTransition{
				Given: given,
				Then:  t.Then,
			})
		}
		if t.Then == given {
			pending = append(pending, closeTransition{
				Given: t.Given,
				Then:  then,
			})
		}
	}
	p.closeTransitions = append(p.closeTransitions, closeTransition{
		Given: given,
		Then:  then,
	})
	for _, t := range pending {
		p.Empty(t.Given, t.Then)
	}
}

func (p *Lexer[T]) Final(given LexerState, then func(start int, text string) T) {
	p.finalStates = append(p.finalStates, finalState[T]{
		Given: given,
		Then:  then,
	})
}

func (p *Lexer[T]) Tokenize(src []byte) ([]T, error) {
	proc := &Stream[T]{
		prog: p,
		src:  src,
		this: make([]bool, p.maxState+1),
		next: make([]bool, p.maxState+1),
	}
	var res []T
	for proc.Next() {
		res = append(res, proc.This())
	}
	if proc.Err() != nil {
		return nil, proc.Err()
	}
	return res, nil
}

func (l *Stream[T]) Err() error {
	return l.err
}

func (l *Stream[T]) Next() bool {
	if l.err != nil {
		return false
	}
	return l.exec()
}

func (l *Stream[T]) This() T {
	return l.tok
}

func (l *Stream[T]) exec() bool {
	pos := l.srcPos
	start := pos
	end := pos
	final := -1
	running := true
	l.this[0] = true

	for running {
		c, n := utf8.DecodeRune(l.src[pos:])
		running = false
		for i := range l.next {
			l.next[i] = false
		}

		l.closeState()
		l.detectFinal(&final, &end, pos)
		l.moveState(&running, c)

		l.this, l.next = l.next, l.this
		pos = pos + n
	}

	if final == -1 {
		return false
	}

	l.tok = l.prog.finalStates[final].Then(start, string(l.src[start:end]))
	l.srcPos = end

	return true
}

func (l *Stream[T]) closeState() {
	for _, op := range l.prog.closeTransitions {
		if !l.this[op.Given] {
			continue
		}
		l.this[op.Then] = true
	}
}

func (l *Stream[T]) detectFinal(final, end *int, pos int) {
	for i, op := range l.prog.finalStates {
		if !l.this[op.Given] {
			continue
		}

		if pos > *end || (pos == *end && i < *final) {
			*end = pos
			*final = i
		}
	}
}

func (l *Stream[T]) moveState(running *bool, c rune) {
	for _, op := range l.prog.moveTransitions {
		if !l.this[op.Given] {
			continue
		}

		if c < op.Min || c > op.Max {
			continue
		}

		l.next[op.Then] = true
		*running = true
	}
}
