package symtab

import (
	"reflect"
	"sort"
	"unsafe"

	"github.com/bobappleyard/cezanne/format/storage"
)

type Symtab struct {
	data []byte
	syms []Symbol
}

type Symbol struct {
	ID uint32
}

type Rewrite struct {
	From, To Symbol
}

func (s Symbol) GTE(t Symbol) bool {
	return s.ID >= t.ID
}

func (t *Symtab) SymbolID(name string) Symbol {
	t.buildSymTable()
	idx := sort.Search(len(t.syms), func(i int) bool {
		return t.SymbolName(t.syms[i]) >= name
	})
	if idx >= len(t.syms) || t.SymbolName(t.syms[idx]) != name {
		return t.addSymbol(idx, name)
	}
	return t.syms[idx]
}

func (t *Symtab) SymbolName(sym Symbol) string {
	end := sym.ID
	for int(end) < len(t.data) && t.data[end] != 0 {
		end++
	}
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&t.data[sym.ID])),
		Len:  int(end - sym.ID),
	}))
}

func (t *Symtab) Merge(u *Symtab) []Rewrite {
	t.buildSymTable()
	u.buildSymTable()

	var rewrites []Rewrite

	for _, from := range u.syms {
		to := t.SymbolID(u.SymbolName(from))

		if from == to {
			continue
		}

		rewrites = append(rewrites, Rewrite{From: from, To: to})
	}

	return rewrites
}

func (t *Symtab) ReadStorage(s *storage.ReadState) error {
	return s.Read(&t.data)
}

func (t Symtab) WriteStorage(s *storage.WriteState) error {
	return s.Write(t.data)
}

func (t *Symtab) buildSymTable() {
	if t.syms != nil {
		return
	}
	inSym := false
	for i, c := range t.data {
		if !inSym {
			t.syms = append(t.syms, Symbol{ID: uint32(i)})
		}
		inSym = c != 0
	}
	sort.Slice(t.syms, func(i, j int) bool {
		return t.SymbolName(t.syms[i]) < t.SymbolName(t.syms[j])
	})
}

func (t *Symtab) addSymbol(at int, s string) Symbol {
	p := Symbol{ID: uint32(len(t.data))}
	t.data = append(t.data, []byte(s)...)
	t.data = append(t.data, 0)
	t.syms = append(t.syms, Symbol{})
	copy(t.syms[at+1:], t.syms[at:])
	t.syms[at] = p
	return p
}
