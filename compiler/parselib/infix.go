package parselib

type InfixParserDef[T any] interface {
	Prec(t Token) int
	Infix(s TokenStream, x T) (T, error)
}

type InfixParser[T any] struct {
	defs map[TokenID]InfixParserDef[T]
}

func (p *InfixParser[T]) Define(tok TokenID, def InfixParserDef[T]) {
	if p.defs == nil {
		p.defs = map[TokenID]InfixParserDef[T]{}
	}
	p.defs[tok] = def
}

func (p *InfixParser[T]) Parse(s TokenStream, prec int, x T) (T, error) {
	var zero T

	for {
		if !s.Next() {
			return x, s.Err()
		}

		t := s.This()
		def := p.defs[t.ID]

		if def == nil || def.Prec(t) <= prec {
			s.Back()
			break
		}

		nx, err := def.Infix(s, x)
		if err != nil {
			return zero, err
		}

		x = nx
	}

	return x, nil
}
