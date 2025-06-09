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

func NewArrayFromTokens(tokens *TokenIterator) (Array, error) {
	size, err := tokens.NextSize(ArrayPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to create array: %w", err)
	}

	output := make(Array, 0, size)

	for range size {
		if value, err := RhelEncode(tokens); err != nil {
			return nil, fmt.Errorf("failed to create array: %w", err)
		} else {
			output = append(output, value)
		}
	}

	return output, nil
}

func (a Array) Size() int {
	size := 0
	for _, i := range a {
		size += i.Size()
	}

	sizeStr := len(strconv.Itoa(len(a)))

	return len(ArrayPrefix) + sizeStr + len(rhelFieldSep) + size
}

func (a Array) Serialize() []byte {
	buf := make([]byte, 0, a.Size())

	buf = append(buf, ArrayPrefix...)
	buf = append(buf, strconv.Itoa(len(a))...)
	buf = append(buf, rhelFieldSep...)

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

func (a Array) isRhelType() {}
