package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

type ByteIterator struct {
	buf    *bufio.Reader
	Offset int
}

func NewByteIteratorFromBytes(data []byte) (iter *ByteIterator) {
	iter = &ByteIterator{
		buf:    bufio.NewReader(bytes.NewReader(data)),
		Offset: 0,
	}

	return
}

func NewByteIteratorFromFile(data *os.File) *ByteIterator {
	return &ByteIterator{buf: bufio.NewReader(data), Offset: 0}
}

func (r *ByteIterator) readBytes(n int) ([]byte, error) {
	buf := make([]byte, n)

	b, err := r.buf.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("expected to read %d bytes, got %d", n, b)
	}

	buf = buf[:b]
	r.Offset += b

	return buf, err
}

func (r *ByteIterator) readByte() (byte, error) {
	r.Offset++

	return r.buf.ReadByte()
}
