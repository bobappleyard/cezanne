package storage

import (
	"errors"
	"fmt"
	"io"
	"reflect"
)

var (
	ErrNotAPointer  = errors.New("not a pointer")
	ErrNotSupported = errors.New("unsupported type")
)

type Reader interface {
	ReadStorage(s *ReadState) error
}

type ReadState struct {
	src      io.ReaderAt
	at, read int64
}

type reference struct {
	Pos, Len uint32
}

// Read a file structure given a root specification.
func Read(src io.ReaderAt, into any) (int64, error) {
	s := &ReadState{
		src: src,
	}
	err := s.Read(into)
	return s.calcRead(), err
}

func (s *ReadState) Read(into any) error {
	if r, ok := into.(Reader); ok {
		return r.ReadStorage(s)
	}
	v := reflect.ValueOf(into)
	if v.Kind() != reflect.Pointer {
		return ErrNotAPointer
	}
	v = v.Elem()
	return s.readValue(v)
}

func (s *ReadState) calcRead() int64 {
	if s.read > s.at {
		return s.read
	}
	return s.at
}

func (s *ReadState) readValue(into reflect.Value) error {
	switch into.Kind() {
	case reflect.Bool:
		var b byte
		if err := s.Read(&b); err != nil {
			return err
		}
		into.SetBool(b != 0)
		return nil
	case reflect.Uint8, reflect.Uint32:
		return s.readUint(into, int(into.Type().Size()))
	case reflect.Int32:
		return s.readInt(into, int(into.Type().Size()))
	case reflect.Struct:
		return s.readStruct(into)
	case reflect.Array:
		return s.readArray(into)
	case reflect.Slice:
		if into.Type().Elem().Kind() == reflect.Uint8 {
			return s.readBytes(into)
		}
		return s.readSlice(into)
	case reflect.String:
		return s.readString(into)
	}
	return fmt.Errorf("%s: %w", into.Type(), ErrNotSupported)
}

func (s *ReadState) readInt(into reflect.Value, size int) error {
	buf := make([]byte, size)
	_, err := s.src.ReadAt(buf, s.at)
	if err != nil {
		return err
	}
	s.at += int64(size)
	var val int64
	for i := 0; i < size; i++ {
		val |= int64(buf[i]) << (i * 8)
	}
	into.SetInt(val)
	return nil
}

func (s *ReadState) readUint(into reflect.Value, size int) error {
	buf := make([]byte, size)
	_, err := s.src.ReadAt(buf, s.at)
	if err != nil {
		return err
	}
	s.at += int64(size)
	var val uint64
	for i := 0; i < size; i++ {
		val |= uint64(buf[i]) << (i * 8)
	}
	into.SetUint(val)
	return nil
}

func (s *ReadState) readStruct(into reflect.Value) error {
	for _, f := range reflect.VisibleFields(into.Type()) {
		if f.PkgPath != "" {
			continue
		}
		if err := s.readValue(into.FieldByIndex(f.Index)); err != nil {
			return err
		}
	}
	return nil
}

func (s *ReadState) readArray(into reflect.Value) error {
	n := into.Len()
	for i := 0; i < n; i++ {
		if err := s.readValue(into.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

func (s *ReadState) readRef() (reference, error) {
	var ref reference
	err := s.Read(&ref)
	return ref, err
}

func (s *ReadState) readBytes(into reflect.Value) error {
	ref, err := s.readRef()
	if err != nil {
		return err
	}
	buf := make([]byte, ref.Len)
	_, err = s.src.ReadAt(buf, int64(ref.Pos))
	if err != nil {
		return err
	}
	into.SetBytes(buf)
	return nil
}

func (s *ReadState) readSlice(into reflect.Value) error {
	ref, err := s.readRef()
	if err != nil {
		return err
	}
	res := reflect.MakeSlice(into.Type(), int(ref.Len), int(ref.Len))
	inner := &ReadState{
		src: s.src,
		at:  int64(ref.Pos),
	}
	for i := 0; i < int(ref.Len); i++ {
		if err := inner.readValue(res.Index(i)); err != nil {
			return err
		}
	}
	into.Set(res)
	if inner.calcRead() > s.read {
		s.read = inner.calcRead()
	}
	return nil
}

func (s *ReadState) readString(into reflect.Value) error {
	ref, err := s.readRef()
	if err != nil {
		return err
	}
	buf := make([]byte, ref.Len)
	_, err = s.src.ReadAt(buf, int64(ref.Pos))
	if err != nil {
		return err
	}
	into.SetString(string(buf))
	return nil
}
