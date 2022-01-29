package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitMethod(t *testing.T) {
	e := New(nil)
	u := &Unit{Code: []byte{
		0, 'c', 'z', 'u',
		1, 0, 0, 0,
		1, 0, 0, 0,
		'c', 'a', 'l', 'l', 0,
		0,
		0, 0, // return register 0
	}}

	_, err := u.exec(e.Process())

	assert.Nil(t, err)
	assert.Equal(t, MethodID(0), u.Methods[0])
}

func TestUnitEntryPoint(t *testing.T) {
	e := New(nil)
	u := &Unit{Code: []byte{
		0, 'c', 'z', 'u',
		1, 0, 0, 0,
		1, 0, 0, 0,
		'c', 'a', 'l', 'l', 0,
		0,
		33, 120, 0, // load literal 120 into register 0
		0, 0, // return register 0
	}}

	res, err := u.exec(e.Process())

	assert.Nil(t, err)
	assert.Equal(t, 120, res.(*intObject).value)
}

// func TestNextArg(t *testing.T) {
// 	for _, test := range []struct {
// 		name  string
// 		size  int
// 		code  []byte
// 		value int
// 	}{
// 		{
// 			name:  "OneByte",
// 			size:  1,
// 			code:  []byte{1},
// 			value: 1,
// 		},
// 		{
// 			name:  "TwoBytes",
// 			size:  2,
// 			code:  []byte{1, 1},
// 			value: 257,
// 		},
// 		{
// 			name:  "FourBytes",
// 			size:  4,
// 			code:  []byte{1, 1, 1, 1},
// 			value: 16843009,
// 		},
// 	} {
// 		t.Run(test.name, func(t *testing.T) {
// 			pos := position{unit: &Unit{Code: test.code}}
// 			res := pos.nextArg(test.size)
// 			assert.Equal(t, test.value, res)
// 		})
// 	}
// }

func TestNextString(t *testing.T) {
	pos := position{unit: &Unit{Code: []byte{'h', 'e', 'l', 'l', 'o', 0, 'w', 'o', 'r', 'l', 'd', 0}}}
	assert.Equal(t, "hello", pos.nextString())
	assert.Equal(t, "world", pos.nextString())
}
