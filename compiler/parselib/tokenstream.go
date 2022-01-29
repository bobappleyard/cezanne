package parselib

import (
	"reflect"
	"unicode/utf8"
	"unsafe"
)

type TokenID int
type StateID int

type Token struct {
	ID         TokenID
	Start, End int
}

type TokenSource interface {
	Next() bool
	This() Token
	Err() error
	Text(t Token) string
}

type TokenStream interface {
	TokenSource
	Back()
}

type MockTokenStream struct {
	source string
	items  []Token
	pos    int
}

type ll1Stream struct {
	src     TokenSource
	last    Token
	useLast bool
}

func LL1(s TokenSource) TokenStream {
	return &ll1Stream{src: s}
}

// Err implements TokenStream
func (s *ll1Stream) Err() error {
	return s.src.Err()
}

// Next implements TokenStream
func (s *ll1Stream) Next() bool {
	if s.useLast {
		s.useLast = false
		return true
	}
	if !s.src.Next() {
		return false
	}
	s.last = s.This()
	return true
}

// Text implements TokenStream
func (s *ll1Stream) Text(t Token) string {
	return s.src.Text(t)
}

// This implements TokenStream
func (s *ll1Stream) This() Token {
	if s.useLast {
		return s.last
	}
	return s.src.This()
}

// Back implements TokenStream
func (s *ll1Stream) Back() {
	s.useLast = true
}

func NewMockTokenStream(source string, items []Token) *MockTokenStream {
	return &MockTokenStream{source, items, -1}
}

// Err implements TokenStream
func (s *MockTokenStream) Err() error {
	return nil
}

// Next implements TokenStream
func (s *MockTokenStream) Next() bool {
	s.pos++
	return s.pos < len(s.items)
}

// Text implements TokenStream
func (s *MockTokenStream) Text(t Token) string {
	return s.source[t.Start:t.End]
}

// This implements TokenStream
func (s *MockTokenStream) This() Token {
	return s.items[s.pos]
}

type LexerProgram struct {
	closeTransitions []closeTransition
	moveTransitions  []moveTransition
	finalStates      []finalState
	maxState         StateID
}

type closeTransition struct {
	Given, Then StateID
}

type moveTransition struct {
	Given, Then StateID
	Min, Max    rune
}

type finalState struct {
	Given   StateID
	TokenID TokenID
}

type Lexer struct {
	prog       *LexerProgram
	src        []byte
	srcPos     int
	this, next []bool
	tok        Token
	err        error
}

func (p *LexerProgram) State() StateID {
	p.maxState++
	return p.maxState
}

func (p *LexerProgram) Rune(given, then StateID, r rune) {
	p.Range(given, then, r, r)
}

func (p *LexerProgram) Range(given, then StateID, min, max rune) {
	p.moveTransitions = append(p.moveTransitions, moveTransition{
		Given: given,
		Then:  then,
		Min:   min,
		Max:   max,
	})
}

func (p *LexerProgram) Empty(given, then StateID) {
	var pending []closeTransition
	for _, t := range p.closeTransitions {
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

func (p *LexerProgram) Final(state StateID, tokenID TokenID) {
	p.finalStates = append(p.finalStates, finalState{
		Given:   state,
		TokenID: tokenID,
	})
}

func (p *LexerProgram) Lexer(src []byte) *Lexer {
	return &Lexer{
		prog: p,
		src:  src,
		this: make([]bool, p.maxState+1),
		next: make([]bool, p.maxState+1),
	}
}

// Err implements TokenStream
func (l *Lexer) Err() error {
	return l.err
}

// Next implements TokenStream
func (l *Lexer) Next() bool {
	if l.err != nil {
		return false
	}
	return l.exec()
}

// Text implements TokenStream
func (l *Lexer) Text(t Token) string {
	buf := (*reflect.SliceHeader)(unsafe.Pointer(&l.src))
	return *((*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: buf.Data + uintptr(t.Start),
		Len:  t.End - t.Start,
	})))
}

// This implements TokenStream
func (l *Lexer) This() Token {
	return l.tok
}

func (l *Lexer) exec() bool {
	pos := l.srcPos

	t := Token{
		Start: l.srcPos,
	}

	running := true
	l.this[0] = true

	for running {
		c, n := utf8.DecodeRune(l.src[pos:])
		running = false
		for i := range l.next {
			l.next[i] = false
		}

		l.closeState()
		l.detectFinal(&t, pos)
		l.moveState(&running, c)

		l.this, l.next = l.next, l.this
		pos = pos + n
	}

	if t.End <= t.Start {
		return false
	}

	l.tok = t
	l.srcPos = t.End

	return true
}

func (l *Lexer) closeState() {
	for _, op := range l.prog.closeTransitions {
		if !l.this[op.Given] {
			continue
		}
		l.this[op.Then] = true
	}
}

func (l *Lexer) detectFinal(t *Token, pos int) {
	for _, op := range l.prog.finalStates {
		if !l.this[op.Given] {
			continue
		}

		if pos > t.End || (pos == t.End && op.TokenID < t.ID) {
			t.End = pos
			t.ID = op.TokenID
		}
	}
}

func (l *Lexer) moveState(running *bool, c rune) {
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
