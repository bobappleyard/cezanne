package anf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	p := Package{
		Data: []byte("hello\x00"),
	}
	assert.Equal(t, "hello", p.String(0))
}
