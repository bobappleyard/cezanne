package storage

import (
	"errors"
	"io"
	"reflect"
)

var (
	ErrNotAPointer  = errors.New("not a pointer")
	ErrNotSupported = errors.New("unsupported type")
)

// Read a file structure given a root specification.
func Read(src io.ReaderAt, into any) error {
	v := reflect.ValueOf(into)
	if v.Kind() != reflect.Pointer {
		return ErrNotAPointer
	}
	v = v.Elem()
	return readValue(src, 0, v)
}

func readValue(src io.ReaderAt, at int64, into reflect.Value) error {
	encoding := getEncoding(into.Type())
	return encoding.read(src, at, into)
}

type writer struct {
	dest io.WriterAt
	tip  int64
}

// Write a file structure given a root specification.
func Write(dest io.WriterAt, from any) error {
	tip := getEncoding(reflect.TypeOf(from)).width()
	w := writer{dest: dest, tip: tip}
	return w.writeValue(0, reflect.ValueOf(from))
}

func (w *writer) writeValue(at int64, from reflect.Value) error {
	encoding := getEncoding(from.Type())
	return encoding.write(w, at, from)
}

func (w *writer) alloc(amt int64) int64 {
	res := w.tip
	w.tip += amt
	return res
}
