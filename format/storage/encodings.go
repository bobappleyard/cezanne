package storage

import (
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/bobappleyard/cezanne/slices"
)

type encoding interface {
	width() int64
	read(from io.ReaderAt, at int64, into reflect.Value) error
	write(w *writer, at int64, from reflect.Value) error
}

type ref struct {
	Start, Len int32
}

var (
	encodings = map[reflect.Type]encoding{}
	lock      sync.RWMutex
)

func getEncoding(t reflect.Type) encoding {
	lock.RLock()
	encoding := encodings[t]
	lock.RUnlock()
	if encoding == nil {
		lock.Lock()
		encoding = buildEncoding(t)
		encodings[t] = encoding
		lock.Unlock()
	}
	return encoding
}

var byteSlice = reflect.TypeOf([]byte(nil))

func buildEncoding(t reflect.Type) encoding {
	switch t.Kind() {
	case reflect.Struct:
		return &structEncoding{
			fields: slices.Map(
				slices.Map(reflect.VisibleFields(t), fieldType),
				buildEncoding,
			),
		}

	case reflect.Bool:
		return &boolEncoding{}

	case reflect.Int32:
		return &intEncoding{bytes: 4}

	case reflect.Uint32:
		return &uintEncoding{bytes: 4}

	case reflect.Slice:
		if t.AssignableTo(byteSlice) {
			return bytesEncoding{}
		}
		return &sliceEncoding{
			elem: buildEncoding(t.Elem()),
		}

	case reflect.String:
		return stringEncoding{}
	}

	panic(fmt.Sprintf("unsupported type %s", t))
}

type boolEncoding struct{}

// read implements encoding
func (*boolEncoding) read(from io.ReaderAt, at int64, into reflect.Value) error {
	var buf [1]byte
	err := readBuffer(from, at, buf[:])
	if err != nil {
		return err
	}
	into.SetBool(buf[0] != 0)
	return nil
}

// width implements encoding
func (*boolEncoding) width() int64 {
	return 1
}

// write implements encoding
func (*boolEncoding) write(w *writer, at int64, from reflect.Value) error {
	var b byte
	if from.Bool() {
		b = 1
	}
	_, err := w.dest.WriteAt([]byte{b}, at)
	return err
}

type structEncoding struct {
	fields []encoding
}

func fieldType(f reflect.StructField) reflect.Type {
	return f.Type
}

// write implements encoding
func (e *structEncoding) write(w *writer, at int64, from reflect.Value) error {
	for i, f := range e.fields {
		err := f.write(w, at, from.Field(i))
		if err != nil {
			return err
		}
		at += f.width()
	}
	return nil
}

// read implements encoding
func (e *structEncoding) read(from io.ReaderAt, at int64, into reflect.Value) error {
	for i, f := range e.fields {
		err := f.read(from, at, into.Field(i))
		if err != nil {
			return err
		}
		at += f.width()
	}
	return nil
}

// width implements encoding
func (e *structEncoding) width() int64 {
	var w int64
	for _, f := range e.fields {
		w += f.width()
	}
	return w
}

type intEncoding struct {
	bytes int64
}

// write implements encoding
func (e *intEncoding) write(w *writer, at int64, from reflect.Value) error {
	buf := make([]byte, e.bytes)
	x := from.Int()
	for i := range buf {
		buf[i] = byte(x >> (i * 8))
	}
	_, err := w.dest.WriteAt(buf, at)
	return err
}

// read implements encoding
func (e *intEncoding) read(from io.ReaderAt, at int64, into reflect.Value) error {
	buf := make([]byte, e.bytes)
	err := readBuffer(from, at, buf)
	if err != nil {
		return err
	}
	var res int64
	for i, b := range buf {
		res |= int64(b) << (8 * i)
	}
	into.SetInt(res)
	return nil
}

// width implements encoding
func (e *intEncoding) width() int64 {
	return e.bytes
}

type uintEncoding struct {
	bytes int64
}

// write implements encoding
func (e *uintEncoding) write(w *writer, at int64, from reflect.Value) error {
	buf := make([]byte, e.bytes)
	x := from.Uint()
	for i := range buf {
		buf[i] = byte(x >> (i * 8))
	}
	_, err := w.dest.WriteAt(buf, at)
	return err
}

// read implements encoding
func (e *uintEncoding) read(from io.ReaderAt, at int64, into reflect.Value) error {
	buf := make([]byte, e.bytes)
	err := readBuffer(from, at, buf)
	if err != nil {
		return err
	}
	var res uint64
	for i, b := range buf {
		res |= uint64(b) << (8 * i)
	}
	into.SetUint(res)
	return nil
}

// width implements encoding
func (e *uintEncoding) width() int64 {
	return e.bytes
}

type sliceEncoding struct {
	elem encoding
}

// write implements encoding
func (e *sliceEncoding) write(w *writer, at int64, from reflect.Value) error {
	width := e.elem.width()
	n := from.Len()
	pos := w.alloc(int64(n) * width)
	err := w.writeValue(at, reflect.ValueOf(ref{
		Start: int32(pos),
		Len:   int32(n),
	}))
	if err != nil {
		return err
	}

	for i := 0; i < n; i++ {
		err := w.writeValue(pos, from.Index(i))
		if err != nil {
			return err
		}
		pos += width
	}

	return nil
}

// read implements encoding
func (e *sliceEncoding) read(from io.ReaderAt, at int64, into reflect.Value) error {
	var pos ref
	err := readValue(from, at, reflect.ValueOf(&pos).Elem())
	if err != nil {
		return err
	}
	into.Set(reflect.MakeSlice(into.Type(), int(pos.Len), int(pos.Len)))

	itemsAt := int64(pos.Start)
	width := e.elem.width()
	for i := 0; i < int(pos.Len); i++ {
		err := readValue(from, itemsAt, into.Index(i))
		if err != nil {
			return err
		}
		itemsAt += width
	}
	return nil
}

// width implements encoding
func (*sliceEncoding) width() int64 {
	return 8
}

type bytesEncoding struct{}

// write implements encoding
func (bytesEncoding) write(w *writer, at int64, from reflect.Value) error {
	buf := from.Bytes()
	pos := w.alloc(int64(len(buf)))
	err := w.writeValue(at, reflect.ValueOf(ref{
		Start: int32(pos),
		Len:   int32(len(buf)),
	}))
	if err != nil {
		return err
	}
	_, err = w.dest.WriteAt(buf, pos)
	return err
}

// read implements encoding
func (bytesEncoding) read(from io.ReaderAt, at int64, into reflect.Value) error {
	var pos ref
	err := readValue(from, at, reflect.ValueOf(&pos).Elem())
	if err != nil {
		return err
	}
	buf := make([]byte, pos.Len)
	err = readBuffer(from, int64(pos.Start), buf)
	if err != nil {
		return err
	}
	into.SetBytes(buf)
	return nil
}

// width implements encoding
func (bytesEncoding) width() int64 {
	return 8
}

type stringEncoding struct{}

// write implements encoding
func (stringEncoding) write(w *writer, at int64, from reflect.Value) error {
	return w.writeValue(at, reflect.ValueOf([]byte(from.String())))
}

// read implements encoding
func (stringEncoding) read(from io.ReaderAt, at int64, into reflect.Value) error {
	var buf []byte
	err := readValue(from, at, reflect.ValueOf(&buf).Elem())
	if err != nil {
		return err
	}
	into.SetString(string(buf))
	return nil
}

// width implements encoding
func (stringEncoding) width() int64 {
	return 8
}

func readBuffer(src io.ReaderAt, at int64, into []byte) error {
	n, err := src.ReadAt(into, at)
	if n < len(into) {
		return err
	}
	return nil
}
