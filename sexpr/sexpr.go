package sexpr

import (
	"fmt"
	"strconv"
)

type Node struct {
	Symbol   string
	Children []Node
}

func Var(x string) Node {
	return Node{Symbol: x}
}

func Bool(x bool) Node {
	if x {
		return Node{Symbol: "#t"}
	}
	return Node{Symbol: "#f"}
}

func String(x string) Node {
	return Node{Symbol: strconv.Quote(x)}
}

func Int(x int) Node {
	return Node{Symbol: strconv.Itoa(x)}
}

func List(xs ...Node) Node {
	if xs == nil {
		xs = []Node{}
	}
	return Node{Children: xs}
}

func Call(f string, xs ...Node) Node {
	return Node{Symbol: f, Children: xs}
}

func DefineFunc(f string, args []string, body ...Node) Node {
	def := append([]Node{Call(f, sliceSelect(args, Var)...)}, body...)
	return Call("define", def...)
}

func DefineVar(v string, x Node) Node {
	return Call("define", Var(v), x)
}

func (n Node) Format(f fmt.State, verb rune) {
	if n.Children != nil {
		f.Write([]byte("("))
	}

	f.Write([]byte(n.Symbol))

	for i, c := range n.Children {
		if i != 0 || n.Symbol != "" {
			f.Write([]byte(" "))
		}
		c.Format(f, verb)
	}

	if n.Children != nil {
		f.Write([]byte(")"))
	}
}

func sliceSelect[T, U any](xs []T, f func(T) U) []U {
	result := make([]U, len(xs))
	for i, x := range xs {
		result[i] = f(x)
	}
	return result
}
