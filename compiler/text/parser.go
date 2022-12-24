package text

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/bobappleyard/cezanne/compiler/stream"
)

type UnexpectedToken struct {
	Token any
}

func (e *UnexpectedToken) Error() string {
	return fmt.Sprintf("unexpected token: %#v", e.Token)
}

func Parse[T, U any](ruleSet any, toks stream.Iter[T]) (U, error) {
	s := &scanner{
		hostType:  reflect.TypeOf(ruleSet),
		tokenType: reflect.TypeOf(new(T)).Elem(),
		rootType:  reflect.TypeOf(new(U)).Elem(),
		types:     map[reflect.Type]*symbol{},
	}
	root := s.scan()

	p := &parser{
		state: [][]item{nil},
	}

	var zero U

	tokVals := stream.Map(toks, func(x T) reflect.Value {
		return reflect.ValueOf(x)
	})
	if err := p.run(root, tokVals); err != nil {
		return zero, err
	}
	if tokVals.Err() != nil {
		return zero, toks.Err()
	}

	rv, err := p.builder().build(root, reflect.ValueOf(ruleSet))
	if err != nil {
		return zero, err
	}

	return rv.Interface().(U), nil
}

type symbol struct {
	// this symbol can be empty
	nullable bool

	// if this is a token rule
	tokenType reflect.Type

	// if this is a nonterminal rule
	predictions []*rule
}

type rule struct {
	// symbol this implements
	implements *symbol

	// array of symbols to match
	deps []*symbol

	// index of method into host
	method int
}

type scanner struct {
	hostType  reflect.Type
	tokenType reflect.Type
	rootType  reflect.Type
	types     map[reflect.Type]*symbol
}

func (s *scanner) scan() *symbol {
	s.ensure(s.rootType)
	s.scanMethods()
	s.markTokenTypes()
	s.markNullableTypes()
	s.fillOutInterfaces()

	return s.types[s.rootType]
}

func (s *scanner) scanMethods() {
	for i := s.hostType.NumMethod() - 1; i >= 0; i-- {
		m := s.hostType.Method(i)
		if !strings.HasPrefix(m.Name, "Parse") {
			continue
		}
		deps := make([]*symbol, m.Type.NumIn()-1)
		for i := m.Type.NumIn() - 1; i >= 1; i-- {
			deps[i-1] = s.ensure(m.Type.In(i))
		}
		produces := s.ensure(m.Type.Out(0))
		produces.predictions = append(produces.predictions, &rule{
			implements: produces,
			deps:       deps,
			method:     m.Index,
		})
	}
}

func (s *scanner) markTokenTypes() {
	for k, v := range s.types {
		if k.AssignableTo(s.tokenType) {
			v.tokenType = k
			continue
		}
	}
}

func (s *scanner) markNullableTypes() {
	var needsWork []*symbol
	symUsers := map[*symbol][]*rule{}

	for _, sym := range s.types {
		for _, r := range sym.predictions {
			for _, s := range r.deps {
				symUsers[s] = append(symUsers[s], r)
			}
			if len(r.deps) == 0 {
				sym.nullable = true
				needsWork = append(needsWork, sym)
			}
		}
	}

	for len(needsWork) != 0 {
		next := needsWork[0]
		needsWork = needsWork[1:]
	nextRule:
		for _, r := range symUsers[next] {
			if r.implements.nullable {
				continue
			}
			for _, s := range r.deps {
				if !s.nullable {
					continue nextRule
				}
			}
			r.implements.nullable = true
			needsWork = append(needsWork, r.implements)
		}
	}
}

func (s *scanner) fillOutInterfaces() {
	var itfs []reflect.Type
	for k := range s.types {
		if k.Kind() != reflect.Interface {
			continue
		}
		itfs = append(itfs, k)
	}
	for len(itfs) != 0 {
		s.fillOutInterface(&itfs, itfs[0])
	}
}

func (s *scanner) fillOutInterface(itfs *[]reflect.Type, todo reflect.Type) {
	if !s.needsFilling(itfs, todo) {
		return
	}
	for k, v := range s.types {
		if k == todo {
			continue
		}
		if !k.AssignableTo(todo) {
			continue
		}
		if k.Kind() == reflect.Interface {
			s.fillOutInterface(itfs, k)
		}
		sym := s.types[todo]
		for _, r := range v.predictions {
			sym.predictions = append(sym.predictions, &rule{
				implements: sym,
				deps:       r.deps,
				method:     r.method,
			})
		}
	}
}

func (s *scanner) needsFilling(itfs *[]reflect.Type, todo reflect.Type) bool {
	set := *itfs
	for i, t := range set {
		if t != todo {
			continue
		}
		copy(set[i:], set[i+1:])
		set = set[:len(set)-1]
		*itfs = set
		return true
	}
	return false
}

func (s *scanner) ensure(key reflect.Type) *symbol {
	if v, ok := s.types[key]; ok {
		return v
	}
	v := new(symbol)
	s.types[key] = v
	return v
}

type parser struct {
	state [][]item
	seen  []reflect.Value
	cur   int
}

type item struct {
	rule     *rule
	position int
	progress int
}

func (p *parser) run(root *symbol, toks stream.Iter[reflect.Value]) error {
	p.state = [][]item{nil}
	p.predict(root)
	for toks.Next() {
		p.seen = append(p.seen, toks.This())
		p.state = append(p.state, nil)

		p.step(toks.This())
		p.cur++
	}
	p.finalStep()
	return p.matches(root)
}

func (p *parser) step(tok reflect.Value) {
	for i := 0; i < len(p.state[p.cur]); i++ {
		item := p.state[p.cur][i]
		next, ok := item.nextSymbol()
		if !ok {
			p.complete(item)
			continue
		}
		if next.tokenType != nil {
			if tok.Type().AssignableTo(next.tokenType) {
				p.scan(item)
			}
			continue
		}
		if next.nullable {
			p.advance(item)
		}
		p.predict(next)
	}
}

func (p *parser) finalStep() {
	for i := 0; i < len(p.state[p.cur]); i++ {
		item := p.state[p.cur][i]
		next, ok := item.nextSymbol()
		if !ok {
			p.complete(item)
			continue
		}
		if next.nullable {
			p.advance(item)
		}
	}
}

func (p *parser) matches(root *symbol) error {
	if len(p.state[len(p.state)-1]) == 0 {
		for i := range p.state[1:] {
			if len(p.state[i+1]) != 0 {
				continue
			}
			return &UnexpectedToken{
				p.seen[i],
			}
		}
	}
	for _, item := range p.state[len(p.state)-1] {
		if item.rule.implements != root {
			continue
		}
		if item.position != 0 {
			continue
		}
		if _, ok := item.nextSymbol(); ok {
			continue
		}
		return nil
	}
	return io.ErrUnexpectedEOF
}

func (p *parser) predict(s *symbol) {
	for _, prediction := range s.predictions {
		p.addToCur(item{
			rule:     prediction,
			position: p.cur,
		})
	}
}

func (p *parser) advance(x item) {
	p.addToCur(x.makeProgress())
}

func (p *parser) scan(x item) {
	p.addToNext(x.makeProgress())
}

func (p *parser) complete(x item) {
	for _, y := range p.state[x.position] {
		next, ok := y.nextSymbol()
		if !ok {
			continue
		}
		if next == x.rule.implements {
			p.addToCur(y.makeProgress())
		}
	}
}

func (p *parser) addToCur(x item) {
	p.addTo(p.cur, x)
}

func (p *parser) addToNext(x item) {
	p.addTo(p.cur+1, x)
}

func (p *parser) addTo(pos int, x item) {
	for _, y := range p.state[pos] {
		if x == y {
			return
		}
	}
	p.state[pos] = append(p.state[pos], x)
}

func (x item) nextSymbol() (*symbol, bool) {
	if x.progress == len(x.rule.deps) {
		return nil, false
	}
	return x.rule.deps[x.progress], true
}

func (x item) makeProgress() item {
	return item{
		rule:     x.rule,
		position: x.position,
		progress: x.progress + 1,
	}
}

var (
	ErrFailedMatch    = errors.New("failed to match")
	ErrAmbiguousParse = errors.New("ambiguous parse")
)

type builder struct {
	state [][]item
	seen  []reflect.Value
}

type span struct {
	item     item
	at       int
	value    reflect.Value
	children []span
}

func (p *parser) builder() *builder {
	flipped := p.flipState()
	for _, s := range flipped {
		sort.Slice(s, func(i, j int) bool {
			im, jm := s[i].rule.method, s[j].rule.method
			if im < jm {
				return true
			}
			if jm < im {
				return false
			}
			return s[i].position < s[j].position
		})
	}
	return &builder{
		state: flipped,
		seen:  p.seen,
	}
}

func (p *parser) flipState() [][]item {
	flipped := make([][]item, len(p.state))
	for i, set := range p.state {
		for _, x := range set {
			if _, ok := x.nextSymbol(); ok {
				continue
			}
			flipped[x.position] = append(flipped[x.position], item{
				rule:     x.rule,
				position: i,
				progress: x.progress,
			})
		}
	}
	return flipped
}

func (b *builder) build(root *symbol, host reflect.Value) (reflect.Value, error) {
	for _, top := range b.state[0] {
		if top.rule.implements != root {
			continue
		}
		if top.position != len(b.seen) {
			continue
		}
		span, ok := b.findSpan(top, 0)
		if !ok {
			return reflect.Value{}, ErrFailedMatch
		}
		return b.buildFromSpan(host, span)
	}
	return reflect.Value{}, ErrFailedMatch
}

func (b *builder) findSpan(x item, at int) (span, bool) {
	children, ok := b.findSpanChildren(x.rule.deps, at, x.position)
	if !ok {
		return span{}, false
	}
	return span{
		item:     x,
		at:       at,
		children: children,
	}, true
}

func (b *builder) buildFromSpan(host reflect.Value, s span) (reflect.Value, error) {
	if s.value.IsValid() {
		return s.value, nil
	}
	args := make([]reflect.Value, len(s.children)+1)
	args[0] = host
	for i, c := range s.children {
		child, err := b.buildFromSpan(host, c)
		if err != nil {
			return reflect.Value{}, err
		}
		args[i+1] = child
	}

	rets := host.Type().Method(s.item.rule.method).Func.Call(args)
	if len(rets) == 2 && !rets[1].IsNil() {
		return reflect.Value{}, rets[1].Interface().(error)
	}
	return rets[0], nil
}

func (b *builder) findSpanChildren(deps []*symbol, at, end int) ([]span, bool) {
	if len(deps) == 0 {
		return nil, at == end
	}
	if deps[0].tokenType != nil {
		return b.tokenSpan(deps, at, end)
	}
	return b.ruleSpan(deps, at, end)
}

func (b *builder) ruleSpan(deps []*symbol, at, end int) ([]span, bool) {
	sym := deps[0]
	for _, found := range b.state[at] {
		if found.rule.implements != sym {
			continue
		}
		next, ok := b.findSpanChildren(deps[1:], found.position, end)
		if !ok {
			continue
		}
		inner, ok := b.findSpan(found, at)
		if !ok {
			continue
		}
		return append([]span{inner}, next...), true
	}
	return nil, false
}

func (b *builder) tokenSpan(deps []*symbol, at, end int) ([]span, bool) {
	sym := deps[0]
	if at >= len(b.seen) {
		return nil, false
	}
	if b.seen[at].Type().AssignableTo(sym.tokenType) {
		next, ok := b.findSpanChildren(deps[1:], at+1, end)
		if ok {
			return append([]span{{
				value: b.seen[at],
				at:    at,
			}}, next...), true
		}
	}
	return nil, false
}
