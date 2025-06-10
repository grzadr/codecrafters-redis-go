package rheltypes

import "fmt"

type Bytes []byte

func (b Bytes) Serialize() (buf []byte) {
	bSize := len(b)
	buf = make([]byte, 0, bSize+3)
	buf = fmt.Append(buf, "$", len(b), "\r\n", b)

	return
}

func (b Bytes) String() string { return string(b) }

func (b Bytes) Integer() (int, error) {
	return 0, nil
}

func (b Bytes) Size() int {
	return len(b)
}

func (b Bytes) First() RhelType {
	return nil
}

func (b Bytes) isRhelType() {}
