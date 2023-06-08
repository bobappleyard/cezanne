package container

import (
	"bytes"
	"errors"
	"io"

	"github.com/bobappleyard/cezanne/format/storage"
	"github.com/bobappleyard/cezanne/format/symtab"
)

var (
	ErrUnknownSegment   = errors.New("unknown segment")
	ErrUnexpectedHeader = errors.New("unexpected header")
)

type Container struct {
	segments map[symtab.Symbol]Segment
	symbols  symtab.Symtab
}

type Segment interface {
	io.ReaderAt
}

func New() *Container {
	c := &Container{
		segments: map[symtab.Symbol]Segment{},
	}
	return c
}

func (c *Container) Symbols() *symtab.Symtab {
	return &c.symbols
}

func (c *Container) Get(segment string) Segment {
	id := c.symbols.SymbolID(segment)
	return c.segments[id]
}

func (c *Container) Put(name string, data Segment) {
	id := c.symbols.SymbolID(name)
	c.segments[id] = data
}

var expectedHead = [...]byte{0, 'C', 'Z', 'p', 1, 1, 0, 0}

func (c *Container) Read(src io.ReaderAt) error {
	var data containerData
	n, err := storage.Read(src, &data)
	if err != nil {
		return err
	}
	if data.Head != expectedHead {
		return ErrUnexpectedHeader
	}

	segments := map[symtab.Symbol]Segment{}
	for _, s := range data.Segments {
		segments[s.ID] = &segmentReader{
			base:  src,
			start: int64(s.Start) + n,
			len:   int64(s.Len),
		}
	}

	c.segments = segments
	c.symbols = data.Symbols

	return nil
}

func (c *Container) Write(dst io.WriterAt) error {
	const bufferSize = 4096

	var bs [bufferSize]byte
	var buf bytes.Buffer
	var data containerData

	data.Head = expectedHead
	data.Symbols = c.symbols

	for id, r := range c.segments {
		start := buf.Len()

		n, p := 0, 0
		var err error
		for err == nil {
			n, err = r.ReadAt(bs[:], int64(p))
			buf.Write(bs[:n])
			p += n
		}
		if err != io.EOF {
			return err
		}

		data.Segments = append(data.Segments, segmentData{
			ID:    id,
			Start: uint32(start),
			Len:   uint32(p),
		})
	}

	n, err := storage.Write(dst, data)
	if err != nil {
		return err
	}

	_, err = dst.WriteAt(buf.Bytes(), n)
	return err
}

type containerData struct {
	Head     [8]byte
	Segments []segmentData
	Symbols  symtab.Symtab
}

type segmentData struct {
	ID         symtab.Symbol
	Start, Len uint32
}

type segmentReader struct {
	base       io.ReaderAt
	start, len int64
}

// ReadAt implements io.ReaderAt.
func (r *segmentReader) ReadAt(p []byte, off int64) (n int, err error) {
	panic("unimplemented")
}
