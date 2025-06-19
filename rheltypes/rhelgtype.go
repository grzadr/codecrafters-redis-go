package rheltypes

import (
	"fmt"
)

var rhelFieldSep = []byte("\r\n")

type RhelType interface {
	isRhelType()
	Serialize() []byte
	String() string
	Integer() (int, error)
	Size() int
	First() RhelType
}

type rhelPrefix string

var (
	UnknownPrefix      = rhelPrefix("")
	SimpleStringPrefix = rhelPrefix("+")
	BulkStringPrefix   = rhelPrefix("$")
	ArrayPrefix        = rhelPrefix("*")
	IntegerPrefix      = rhelPrefix(":")
	// rhelPrefixIndex    = []rhelPrefix{
	// 	SimpleStringPrefix,
	// 	BulkStringPrefix,
	// 	ArrayPrefix,
	// 	IntegerPrefix,
	// }.
)

func NewRhelPrefix(p string) rhelPrefix {
	switch rhelPrefix(p) {
	case SimpleStringPrefix, BulkStringPrefix, ArrayPrefix, IntegerPrefix:
		return rhelPrefix(p)
	default:
		return UnknownPrefix
	}
}

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
	switch p := iter.LastToken.Prefix; p {
	case ArrayPrefix:
		return NewArrayFromTokens(iter)
	case SimpleStringPrefix:
		return NewSimpleStringFromTokens(iter)
	case BulkStringPrefix:
		bytes, err := NewBytesFromTokens(iter)
		if err != nil {
			return nil, err
		}

		return NewBulkStringFromTokens(iter)
	case IntegerPrefix:
		return NewIntegerFromTokens(iter)
	default:
		return nil, fmt.Errorf("unknown prefix %s", string(p))
	}
}
