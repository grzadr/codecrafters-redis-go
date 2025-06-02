package rheltypes

import (
	"fmt"
	"strconv"
)

type RhelType interface {
	isRhelType()
	Serialize() []byte
}

type SimpleString string

func (s SimpleString) isRhelType() {}

func (s SimpleString) Serialize() []byte {
	buf := make([]byte, 0, len(s)+3)

	return fmt.Appendf(buf, "+%s\r\n", s)
}

type BulkString string

func (s BulkString) isRhelType() {}

func (s BulkString) Serialize() []byte {
	sizeStr := strconv.Itoa(len(s))
	buf := make([]byte, 0, len(s)+len(sizeStr)+5)

	return fmt.Appendf(buf, "$%s\r\n%s\r\n", sizeStr, s)
}
