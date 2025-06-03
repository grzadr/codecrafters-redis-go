package rheltypes

import (
	"bytes"
	"fmt"
	"strconv"
)

type SimpleString string

func NewSimpleString(content []byte) (SimpleString, error) {
	after, found := bytes.CutPrefix(content, SimpleStringPrefix)
	if !found {
		return "", PrefixError{Content: content, Prefix: SimpleStringPrefix}
	}
	return SimpleString(string(after)), nil
}

func (s SimpleString) isRhelType() {}

func (s SimpleString) Size() int {
	size := len(s)
	sizeStr := strconv.Itoa(size)

	return len(
		SimpleStringPrefix,
	) + len(
		sizeStr,
	) + len(
		rhelFieldSep,
	) + size + len(
		rhelFieldSep,
	)
}

func (s SimpleString) Serialize() []byte {
	buf := make([]byte, 0, s.Size())

	return fmt.Appendf(buf, "+%s\r\n", s)
}

func (s SimpleString) String() string {
	return string(s)
}

type BulkString string

func NewBulkString(content []byte) (BulkString, error) {
	data, found := bytes.CutPrefix(content, BulkStringPrefix)
	if !found {
		return "", PrefixError{Content: content, Prefix: BulkStringPrefix}
	}
	_, data, found = bytes.Cut(data, rhelFieldSep)

	if !found {
		return "", ContentError(content)
	}

	data, _ = bytes.CutSuffix(data, rhelFieldSep)

	return BulkString(data), nil
}

func (s BulkString) isRhelType() {}

func (s BulkString) Size() int {
	return len(s) + len(SimpleStringPrefix) + 2
}

func (s BulkString) Serialize() []byte {
	buf := make([]byte, 0, s.Size())

	return fmt.Appendf(buf, "$%s\r\n%s\r\n", strconv.Itoa(len(s)), s)
}

func (s BulkString) String() string {
	return string(s)
}
