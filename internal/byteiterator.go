package internal

import (
	"bytes"
	"io"
	"os"
)

type ByteIterator struct {
	Input  io.Reader
	Offset int
}

func NewByteIteratorFromBytes(data []byte) *ByteIterator {
	return &ByteIterator{Input: bytes.NewReader(data), Offset: 0}
}

func NewByteIteratorFromFile(data *os.File) *ByteIterator {
	return &ByteIterator{Input: data, Offset: 0}
}
