package rheltypes

import (
	"fmt"
)

var rhelFieldDelim = []byte("\r\n")

type RhelType interface {
	isRhelType()
	Serialize() []byte
	String() string
	Integer() (int, error)
	Size() int
	First() RhelType
	TypeName() string
}

type rhelPrefix string

var (
	ArrayPrefix        = rhelPrefix("*")
	BulkStringPrefix   = rhelPrefix("$")
	ErrorPrefix        = rhelPrefix("-")
	IntegerPrefix      = rhelPrefix(":")
	SimpleStringPrefix = rhelPrefix("+")
	UnknownPrefix      = rhelPrefix("")
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
	token, err := iter.NextToken()
	if err != nil {
		return nil, fmt.Errorf("failed to read next token: %w", err)
	}

	if iter.IsDone() {
		return nil, nil
	}

	switch token.Prefix {
	case ArrayPrefix:
		return NewArrayFromTokens(token, iter)
	case SimpleStringPrefix:
		return NewSimpleStringFromTokens(token)
	case BulkStringPrefix:
		return NewBulkStringFromTokens(token, iter)
	case IntegerPrefix:
		return NewIntegerFromTokens(token)
	default:
		return nil, fmt.Errorf("unsupported prefix %q", token)
	}
}
