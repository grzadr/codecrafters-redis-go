package rheltypes

import (
	"fmt"
	"strconv"
)

type SimpleString string

func NewSimpleStringFromTokens(token Token) (SimpleString, error) {
	return SimpleString(token.Data), nil
}

func (s SimpleString) Size() int {
	return len(s) + len(SimpleStringPrefix) + len(rhelFieldDelim)
}

func (s SimpleString) Serialize() []byte {
	buf := make([]byte, 0, s.Size())

	return fmt.Appendf(buf, "+%s\r\n", s)
}

func (s SimpleString) String() string {
	return string(s)
}

func (s SimpleString) First() RhelType {
	return s
}

func (s SimpleString) Integer() (num int, err error) {
	return strconv.Atoi(string(s))
}

func (s SimpleString) isRhelType() {}

type BulkString struct {
	Length     int
	Text       []byte
	terminated bool
}

func NewBulkString(str string) (bs BulkString) {
	bs.Length = len(str)
	bs.Text = []byte(str)
	bs.terminated = true

	return
}

func NewBulkStringFromTokens(
	token Token,
	iter *TokenIterator,
) (bs BulkString, err error) {
	bs.Length, err = token.AsSize()
	if err != nil {
		return bs, fmt.Errorf(
			"failed to read bul string size from %q: %w",
			token,
			err,
		)
	}

	bs.Text, err = iter.readBytes(bs.Length)
	if err != nil {
		return bs, fmt.Errorf(
			"failed to read %d bulk content bytes: %w",
			bs.Length,
			err,
		)
	}

	bs.terminated, err = iter.skipDelim(rhelFieldDelim)
	if err != nil {
		return bs, fmt.Errorf("failed to read following field delimiter: %w")
	}

	return
}

func NewNullBulkString() BulkString {
	return BulkString{Length: -1}
}

func (s BulkString) Size() int {
	sizeStr := strconv.Itoa(s.Length)

	return len(
		SimpleStringPrefix,
	) + len(
		sizeStr,
	) + len(
		rhelFieldDelim,
	) + max(s.Length, 0) + len(
		rhelFieldDelim,
	)
}

func (s BulkString) Serialize() []byte {
	buf := make([]byte, 0, s.Size())

	buf = fmt.Appendf(buf, "%s%s\r\n", BulkStringPrefix, strconv.Itoa(s.Length))

	if s.Length > -1 {
		buf = fmt.Appendf(buf, "%s\r\n", s.Text)
	}

	return buf
}

func (s BulkString) String() string {
	return string(s.Text)
}

func (s BulkString) First() RhelType {
	return s
}

func (s BulkString) Integer() (num int, err error) {
	return strconv.Atoi(string(s.Text))
}
func (s BulkString) isRhelType() {}
