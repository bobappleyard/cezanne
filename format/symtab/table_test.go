package symtab

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/bobappleyard/cezanne/format/storage"
	"github.com/bobappleyard/cezanne/util/assert"
	"github.com/bobappleyard/cezanne/util/slices"
)

func TestAddSymbols(t *testing.T) {
	var tab Symtab
	assert.Equal(t, tab.SymbolID("symtab"), Symbol{0})
	assert.Equal(t, tab.SymbolID("symtab"), Symbol{0})
	assert.Equal(t, tab.SymbolID("types/v1"), Symbol{7})
	assert.Equal(t, tab.SymbolID("tiles/v1"), Symbol{16})
	t.Log(tab)
}

func generateSymbol(r *rand.Rand) string {
	bs := make([]byte, r.Intn(100))
	for i := range bs {
		bs[i] = byte(r.Intn(96)) + 32
	}
	return string(bs)
}

func TestSymbolName(t *testing.T) {
	t.Log(generateSymbol(rand.New(rand.NewSource(0))))
	var tab Symtab
	if err := quick.CheckEqual(
		func(s string) string { return s },
		func(s string) string { return tab.SymbolName(tab.SymbolID(s)) },
		&quick.Config{Values: func(v []reflect.Value, r *rand.Rand) {
			for i := range v {
				v[i] = reflect.ValueOf(generateSymbol(r))
			}
		}},
	); err != nil {
		t.Error(err)
	}
}

func TestMerge(t *testing.T) {
	var tab Symtab
	tab.SymbolID("test")
	var mtab Symtab
	from := mtab.SymbolID("test2")

	rw := tab.Merge(&mtab)
	to := tab.SymbolID("test2")

	assert.True(t, len(slices.Filter(rw, func(rw Rewrite) bool {
		return rw.From == from && rw.To == to
	})) == 1)
}

type testBuf struct {
	buf []byte
}

// WriteAt implements io.WriterAt
func (b *testBuf) WriteAt(p []byte, off int64) (int, error) {
	n := int(off) + len(p)
	if n > len(b.buf) {
		b.buf = append(b.buf, make([]byte, n-len(b.buf))...)
	}
	copy(b.buf[off:], p)
	return len(p), nil
}

func (b *testBuf) ReadAt(p []byte, off int64) (int, error) {
	copy(p, b.buf[off:])
	return len(p), nil
}

func TestIO(t *testing.T) {
	var tab Symtab
	for c := 'a'; c <= 'z'; c++ {
		tab.SymbolID(string([]byte{byte(c)}))
	}

	sym := tab.SymbolID("sym")

	var buf testBuf
	_, err := storage.Write(&buf, tab)

	assert.Nil(t, err)
	t.Log(buf)
	t.Log(tab)

	var rtab Symtab
	_, err = storage.Read(&buf, &rtab)
	assert.Nil(t, err)

	t.Log(rtab)
	assert.Equal(t, rtab.data, tab.data)
	assert.Equal(t, rtab.SymbolID("sym"), sym)
}
