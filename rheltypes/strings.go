package rheltypes

import (
	"fmt"
	"strconv"
)

type SimpleString string

func NewSimpleStringFromTokens(iter *TokenIterator) (SimpleString, error) {
	first, ok := iter.Read(1)
	if !ok {
		return "", fmt.Errorf("failed to read value token")
	}

	value, found := first[0].CutPrefix(SimpleStringPrefix)
	if !found {
		return "", NewPrefixError(SimpleStringPrefix, rhelPrefix(first[0]))
	}

	return SimpleString(string(value)), nil
}

func (s SimpleString) Size() int {
	return len(s) + len(SimpleStringPrefix) + len(rhelFieldSep)
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
	Length int
	Text   string
}

func NewBulkString(str string) (bs BulkString) {
	bs.Length = len(str)
	bs.Text = str

	return
}

func NewBulkStringFromTokens(tokens *TokenIterator) (bs BulkString, err error) {
	bs.Length, err = tokens.NextSize(BulkStringPrefix)
	if err != nil {
		return bs, fmt.Errorf("failed to create bulk string: %w", err)
	}

	valueToken, ok := tokens.Next()
	if !ok {
		return bs, fmt.Errorf(
			"failed to create bulk string: failed to read value token",
		)
	}

	bs.Text = string(valueToken)

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
		rhelFieldSep,
	) + max(s.Length, 0) + len(
		rhelFieldSep,
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
