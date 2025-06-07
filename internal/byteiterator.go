package internal

type ByteIterator struct {
	Data   []byte
	Offset int
}

func NewByteIterator(data []byte) *ByteIterator {
	return &ByteIterator{Data: data, Offset: 0}
}
