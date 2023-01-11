package storage

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	var file struct {
		Name  string
		Items []struct {
			ID     int32
			Offset int32
		}
	}

	s := bytes.NewReader([]byte{
		16, 0, 0, 0, 5, 0, 0, 0,
		21, 0, 0, 0, 2, 0, 0, 0,
		'h', 'e', 'l', 'l', 'o',
		12, 0, 1, 0, 0, 0, 0, 0,
		37, 0, 0, 2, 0, 0, 0, 0, 0,
	})

	err := Read(s, &file)
	assert.Nil(t, err)

	assert.Equal(t, "hello", file.Name)
	assert.Equal(t, 65548, int(file.Items[0].ID))
	assert.Equal(t, 33554469, int(file.Items[1].ID))
}

type testWriter struct {
	buf []byte
}

// WriteAt implements io.WriterAt
func (w *testWriter) WriteAt(p []byte, off int64) (int, error) {
	n := int(off) + len(p)
	if n > len(w.buf) {
		w.buf = append(w.buf, make([]byte, n-len(w.buf))...)
	}
	copy(w.buf[off:], p)
	return len(p), nil
}

func TestWrite(t *testing.T) {
	type bs []byte
	type item struct {
		ID     int32
		Offset int32
		Bs     bs
	}
	type file struct {
		Name  string
		Items []item
	}

	var d testWriter
	e := file{
		Name:  "hello",
		Items: []item{{1, 10000, bs{}}, {2678, 4, bs{1}}},
	}

	err := Write(&d, e)
	assert.Nil(t, err)

	var f file
	err = Read(bytes.NewReader(d.buf), &f)
	assert.Nil(t, err)

	assert.Equal(t, e, f)
}
