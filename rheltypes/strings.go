package rheltypes

import (
	"fmt"
	"slices"
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

func (s SimpleString) TypeName() string {
	return "string"
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

func NewBulkStringFromBytes(b []byte) (bs BulkString) {
	bs.Length = len(b)
	bs.Text = b
	bs.terminated = false

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
		return bs, fmt.Errorf(
			"failed to read following field delimiter: %w",
			err,
		)
	}

	return bs, err
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
	switch s.Length {
	case -1:
		return slices.Concat(
			[]byte(BulkStringPrefix),
			[]byte("-1"),
			rhelFieldDelim,
		)
	case 0:
		return slices.Concat(
			[]byte(BulkStringPrefix),
			[]byte("0"),
			rhelFieldDelim,
			rhelFieldDelim,
		)
	default:
		buf := make([]byte, 0, s.Size())
		buf = fmt.Appendf(
			buf,
			"%s%s\r\n",
			BulkStringPrefix,
			strconv.Itoa(s.Length),
		)

		if s.terminated {
			buf = fmt.Appendf(buf, "%s\r\n", s.Text)
		} else {
			buf = slices.Concat(buf, s.Text)
		}

		return buf
	}
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

func (s BulkString) IsTerminated() bool {
	return s.terminated
}

func (s BulkString) TypeName() string {
	return "string"
}

func (s BulkString) isRhelType() {}
