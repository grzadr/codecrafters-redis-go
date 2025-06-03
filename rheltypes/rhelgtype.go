package rheltypes

import (
	"fmt"
)

var rhelFieldSep = []byte("\r\n")

type RhelType interface {
	isRhelType()
	Serialize() []byte
	String() string
	Size() int
	First() RhelType
}

type rhelPrefix string

var (
	SimpleStringPrefix = rhelPrefix("+")
	BulkStringPrefix   = rhelPrefix("$")
	ArrayPrefix        = rhelPrefix("*")
	rhelPrefixIndex    = []rhelPrefix{
		SimpleStringPrefix,
		BulkStringPrefix,
		ArrayPrefix,
	}
)

type PrefixError struct {
	expected rhelPrefix
	detected rhelPrefix
}

func (e PrefixError) Error() string {
	return fmt.Sprintf("expected prefix %q, got %q",
		e.expected, e.detected)
}

func NewPrefixError(expected, detected rhelPrefix) error {
	return PrefixError{expected: expected, detected: detected}
}

func RhelEncode(iter *TokenIterator) (RhelType, error) {
	switch p := iter.Current().Prefix(); p {
	case ArrayPrefix:
		return NewArray(iter)
	case SimpleStringPrefix:
		return NewSimpleString(iter)
	case BulkStringPrefix:
		return NewBulkString(iter)
	default:
		return nil, fmt.Errorf("unknown prefix %s", string(p))
	}
}
