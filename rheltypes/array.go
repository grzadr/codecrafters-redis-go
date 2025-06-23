package rheltypes

import (
	"fmt"
	"strconv"
	"strings"
)

type Array []RhelType

func NewArrayFromStrings(values []string) (a Array) {
	a = make(Array, len(values))

	for i, v := range values {
		a[i] = NewBulkString(v)
	}

	return
}

func NewArrayFromTokens(token Token, iter *TokenIterator) (a Array, err error) {
	length, err := token.AsSize()
	if err != nil {
		return nil, fmt.Errorf("failed to init bul string: %w", err)
	}

	a = make(Array, 0, length)

	for range length {
		value, err := RhelEncode(iter)
		if err != nil {
			return nil, fmt.Errorf("failed to create array: %w", err)
		}

		a = append(a, value)
	}

	return
}

func (a Array) Size() int {
	size := 0
	for _, i := range a {
		size += i.Size()
	}

	sizeStr := len(strconv.Itoa(len(a)))

	return len(ArrayPrefix) + sizeStr + len(rhelFieldDelim) + size
}

func (a Array) Serialize() []byte {
	buf := make([]byte, 0, a.Size())

	buf = append(buf, ArrayPrefix...)
	buf = append(buf, strconv.Itoa(len(a))...)
	buf = append(buf, rhelFieldDelim...)

	for _, element := range a {
		buf = append(buf, element.Serialize()...)
	}

	return buf
}

func (a Array) String() string {
	buf := make([]string, 0, len(a))
	for _, i := range a {
		buf = append(buf, i.String())
	}

	return strings.Join(buf, ", ")
}

func (a Array) First() RhelType {
	if len(a) == 0 {
		return nil
	}

	return a[0]
}

func (a Array) At(index int) RhelType {
	if index < 0 {
		return a.At(len(a) + index)
	} else if index < len(a) {
		return a[index]
	} else {
		return nil
	}
}

func (a Array) Integer() (int, error) { return 0, nil }

func (a Array) Cmd() string {
	return strings.ToUpper(a.First().String())
}

func (a Array) TypeName() string {
	return "array"
}

func (a Array) isRhelType() {}
