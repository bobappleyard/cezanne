package sexpr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	for _, test := range []struct {
		name string
		in   Node
		out  string
	}{
		{
			name: "Integer",
			in:   Int(234),
			out:  `234`,
		},
		{
			name: "String",
			in:   String("hello"),
			out:  `"hello"`,
		},
		{
			name: "Var",
			in:   Var("x"),
			out:  `x`,
		},
		{
			name: "EmptyList",
			in:   List(),
			out:  `()`,
		},
		{
			name: "List",
			in:   List(Int(1)),
			out:  `(1)`,
		},
		{
			name: "Call",
			in:   Call("test-function", Int(1), Int(2)),
			out:  `(test-function 1 2)`,
		},
		{
			name: "Define",
			in:   DefineFunc("test-function", []string{"a"}, Int(1)),
			out:  `(define (test-function a) 1)`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.out, fmt.Sprint(test.in))
		})
	}
}
