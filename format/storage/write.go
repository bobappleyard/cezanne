package storage

import (
	"fmt"
	"io"
	"reflect"
)

type Writer interface {
	WriteStorage(s *WriteState) error
}

type WriteState struct {
	dest    io.WriterAt
	at      int64
	pending []work
}

type work struct {
	at   int64
	data Writer
}

// Write a file structure given a root specification.
func Write(dest io.WriterAt, from any) (int64, error) {
	s := &WriteState{
		dest: dest,
	}
	if err := s.Write(from); err != nil {
		return 0, err
	}
	for len(s.pending) != 0 {
		work := s.pending
		s.pending = nil
		for _, r := range work {
			at := s.at
			s.at = r.at
			if err := s.Write(uint32(at)); err != nil {
				return 0, err
			}
			s.at = at
			if err := s.Write(r.data); err != nil {
				return 0, err
			}
		}
	}
	return s.at, nil
}

func (s *WriteState) Write(from any) error {
	if w, ok := from.(Writer); ok {
		return w.WriteStorage(s)
	}
	return s.writeValue(reflect.ValueOf(from))
}

func (s *WriteState) writeValue(from reflect.Value) error {
	switch from.Kind() {
	case reflect.Bool:
		fmt.Println(s.at, from.Bool())
		if from.Bool() {
			return s.Write(byte(1))
		}
		return s.Write(byte(0))
	case reflect.Int32:
		return s.writeInt(from, int(from.Type().Size()))
	case reflect.Uint8, reflect.Uint32:
		return s.writeUint(from, int(from.Type().Size()))
	case reflect.Struct:
		return s.writeStruct(from)
	case reflect.Array:
		return s.writeArray(from)
	case reflect.Slice:
		if from.Type().Elem().Kind() == reflect.Uint8 {
			return s.writeBytes(from)
		}
		return s.writeSlice(from)
	case reflect.String:
		return s.writeString(from)
	}
	return fmt.Errorf("%s: %w", from.Type(), ErrNotSupported)
}

func (s *WriteState) writeInt(from reflect.Value, size int) error {
	val := from.Int()
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = byte(val >> (i * 8))
	}
	_, err := s.dest.WriteAt(buf, s.at)
	if err != nil {
		return err
	}
	s.at += int64(size)
	return nil
}

func (s *WriteState) writeUint(from reflect.Value, size int) error {
	val := from.Uint()
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = byte(val >> (i * 8))
	}
	_, err := s.dest.WriteAt(buf, s.at)
	if err != nil {
		return err
	}
	s.at += int64(size)
	return nil
}

func (s *WriteState) writeStruct(from reflect.Value) error {
	for _, f := range reflect.VisibleFields(from.Type()) {
		if f.PkgPath != "" {
			continue
		}
		if err := s.writeValue(from.FieldByIndex(f.Index)); err != nil {
			return err
		}
	}
	return nil
}

func (s *WriteState) writeArray(from reflect.Value) error {
	n := from.Len()
	for i := 0; i < n; i++ {
		if err := s.writeValue(from.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

func (s *WriteState) writeSlice(from reflect.Value) error {
	at := s.at
	if err := s.Write(reference{Len: uint32(from.Len())}); err != nil {
		return err
	}
	s.pending = append(s.pending, work{at: at, data: &sliceData{from}})
	return nil
}

type sliceData struct {
	data reflect.Value
}

func (d *sliceData) WriteStorage(s *WriteState) error {
	n := d.data.Len()
	for i := 0; i < n; i++ {
		if err := s.writeValue(d.data.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

func (s *WriteState) writeBytes(from reflect.Value) error {
	at := s.at
	if err := s.Write(reference{Len: uint32(from.Len())}); err != nil {
		return err
	}
	s.pending = append(s.pending, work{at: at, data: &bytesData{from.Bytes()}})
	return nil
}

func (s *WriteState) writeString(from reflect.Value) error {
	at := s.at
	if err := s.Write(reference{Len: uint32(from.Len())}); err != nil {
		return err
	}
	s.pending = append(s.pending, work{at: at, data: &bytesData{[]byte(from.String())}})
	return nil
}

type bytesData struct {
	data []byte
}

func (d *bytesData) WriteStorage(s *WriteState) error {
	_, err := s.dest.WriteAt(d.data, s.at)
	if err != nil {
		return err
	}
	s.at += int64(len(d.data))
	return nil
}
