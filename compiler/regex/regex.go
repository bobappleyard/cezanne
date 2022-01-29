package regex

import "github.com/bobappleyard/cezanne/compiler/parselib"

func Regex(prog *parselib.LexerProgram, tok parselib.TokenID, re string) error {
	end := prog.State()
	prog.Final(end, tok)

	var p parser
	p.init()

	l := parselib.LL1(regexProg.Lexer([]byte(re)))
	e, err := p.parse(l, 0)
	if err != nil {
		return err
	}

	compileRegex(prog, 0, end, e)
	return nil
}

func compileRegex(p *parselib.LexerProgram, from, to parselib.StateID, e expr) {
	switch e.typ {
	case charExpr:
		p.Range(from, to, e.min, e.max)

	case seqExpr:
		for _, e := range e.sub {
			mid := p.State()
			compileRegex(p, from, mid, e)
			from = mid
		}
		p.Empty(from, to)

	case choiceExpr:
		for _, e := range e.sub {
			compileRegex(p, from, to, e)
		}

	case repeatExpr:
		// kleene closure
		s1, s2 := p.State(), p.State()
		p.Empty(from, s1)
		p.Empty(s2, s1)
		p.Empty(s2, to)
		compileRegex(p, s1, s2, e.sub[0])

	}
}
