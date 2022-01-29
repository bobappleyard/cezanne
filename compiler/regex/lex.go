package regex

import "github.com/bobappleyard/cezanne/compiler/parselib"

const (
	invalidToken = iota
	chOpToken
	chClToken
	rangeToken
	inverseToken
	grOpToken
	grClToken
	starToken
	optionToken
	repeatToken
	choiceToken
	anyToken
	escToken
	charToken
)

var regexProg parselib.LexerProgram

func init() {
	singleCharOp := func(r rune, token parselib.TokenID) {
		s := regexProg.State()
		regexProg.Rune(0, s, r)
		regexProg.Final(s, token)
	}

	singleCharOp('[', chOpToken)
	singleCharOp(']', chClToken)
	singleCharOp('-', rangeToken)
	singleCharOp('^', inverseToken)
	singleCharOp('(', grOpToken)
	singleCharOp(')', grClToken)
	singleCharOp('*', starToken)
	singleCharOp('?', optionToken)
	singleCharOp('+', repeatToken)
	singleCharOp('|', choiceToken)
	singleCharOp('.', anyToken)

	escMid := regexProg.State()
	escEnd := regexProg.State()
	anyEnd := regexProg.State()

	regexProg.Range(0, anyEnd, ' ', '~')
	regexProg.Rune(0, escMid, '\\')
	regexProg.Range(escMid, escEnd, ' ', '~')

	regexProg.Final(escEnd, escToken)
	regexProg.Final(anyEnd, charToken)
}
