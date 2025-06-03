package rheltypes

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

var rhelFieldSep = []byte("\r\n")

type RhelType interface {
	isRhelType()
	Serialize() []byte
	String() string
	Size() int
}

type rhelPrefix []byte

var (
	SimpleStringPrefix = rhelPrefix("+")
	BulkStringPrefix   = rhelPrefix("$")
	ArrayPrefix        = rhelPrefix("*")
)

func RhelSerialize(content []byte) (RhelType, error) {
	switch content[0] {
	case '*':
		return NewArray(content)
	case '+':
		return NewSimpleString(content)
	case '$':
		return NewBulkString(content)
	default:
		return nil, fmt.Errorf(
			"unknown type '%s': %s",
			string(content[0]),
			string(content),
		)
	}
}

type PrefixError struct {
	Content []byte
	Prefix  rhelPrefix
}

func (e PrefixError) Error() string {
	return fmt.Sprintf("missing expected prefix %q in %q:\n%s",
		e.Prefix, e.Content, hex.Dump(e.Content))
}

func NewPrefixError(content []byte, prefix rhelPrefix) error {
	return PrefixError{Content: content, Prefix: prefix}
}

type ContentError []byte

func (e ContentError) Error() string {
	return fmt.Sprintf("malformed content %q:\n%s",
		string(e), hex.Dump(e))
}

func NewContentError(content []byte) error {
	return ContentError(content)
}

func cutRhelPrefix(
	content []byte,
	prefix rhelPrefix,
) (clean []byte, err error) {
	var found bool
	if clean, found = bytes.CutPrefix(content, prefix); !found {
		err = NewPrefixError(content, prefix)
	}
	return
}
