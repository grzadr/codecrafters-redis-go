package rheltypes

import (
	"fmt"
	"strconv"
)

type SimpleString string

func NewSimpleString(iter *TokenIterator) (SimpleString, error) {
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

func (s SimpleString) First() RhelType {
	return s
}

type BulkString string

func NewBulkString(tokens *TokenIterator) (s BulkString, err error) {
	size, err := tokens.NextSize(BulkStringPrefix)
	if err != nil {
		return "", fmt.Errorf("failed to create bulk string: %w", err)
	}

	if valueToken, ok := tokens.Next(); !ok {
		return "", fmt.Errorf(
			"failed to create bulk string: failed to read value token",
		)
	} else {
		s = BulkString(valueToken)
	}

	if sSize := len(s); size != sSize {
		return "", fmt.Errorf(
			"failed to create bulk string: expected %q to have size %d, has %d",
			s,
			size,
			sSize,
		)
	}

	return
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

func (s BulkString) First() RhelType {
	return s
}
