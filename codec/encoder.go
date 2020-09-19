package codec

import (
	"encoding/binary"
	"io"
)

type Serializable interface {
	Serialize(w *EncodingWriter) error
}

type Encodable interface {
	Encode() ([]byte, error)
}

type EncodingWriter struct {
	w       io.Writer
	n       int
	Scratch [32]byte
}

func NewEncodingWriter(w io.Writer) *EncodingWriter {
	return &EncodingWriter{w: w, n: 0}
}

// How many bytes were written to the underlying io.Writer before ending encoding (for handling errors)
func (ew *EncodingWriter) Written() int {
	return ew.n
}

// Write writes len(p) bytes from p fully to the underlying accumulated buffer.
func (ew *EncodingWriter) Write(p []byte) error {
	n := 0
	for n < len(p) {
		d, err := ew.w.Write(p[n:])
		ew.n += d
		if err != nil {
			return err
		}
		n += d
	}
	return nil
}

// Write a single byte to the buffer.
func (ew *EncodingWriter) WriteByte(v byte) error {
	ew.Scratch[0] = v
	return ew.Write(ew.Scratch[0:1])
}

// Writes an offset for an element
func (ew *EncodingWriter) WriteOffset(prevOffset uint64, elemLen uint64) (offset uint64, err error) {
	if prevOffset >= (uint64(1) << 32) {
		panic("cannot write offset with invalid previous offset")
	}
	if elemLen >= (uint64(1) << 32) {
		panic("cannot write offset with invalid element size")
	}
	offset = prevOffset + elemLen
	if offset >= (uint64(1) << 32) {
		panic("offset too large, not uint32")
	}
	binary.LittleEndian.PutUint32(ew.Scratch[0:4], uint32(offset))
	err = ew.Write(ew.Scratch[0:4])
	return
}

func (ew *EncodingWriter) WriteUint16(v uint16) error {
	binary.LittleEndian.PutUint16(ew.Scratch[0:2], v)
	return ew.Write(ew.Scratch[0:2])
}

func (ew *EncodingWriter) WriteUint32(v uint32) error {
	binary.LittleEndian.PutUint32(ew.Scratch[0:4], v)
	return ew.Write(ew.Scratch[0:4])
}

func (ew *EncodingWriter) WriteUint64(v uint64) error {
	binary.LittleEndian.PutUint64(ew.Scratch[0:8], v)
	return ew.Write(ew.Scratch[0:8])
}
