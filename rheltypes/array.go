package rheltypes

import (
	"fmt"
	"strconv"
	"strings"
)

type Array []RhelType

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

// func (a Array) Slice(from, to int) (s Array, err error) {
// 	if from < 0 || from > to {
// 		return nil, fmt.Errorf(
// 			"invalid slice range: from %d must be >= 0 and < %d",
// 			from,
// 			to,
// 		)
// 	}

// 	s = make(Array, to-from)

// 	if len(a) - from > len(s) {
// 		return nil, fmt.Errorf("")
// 	}

// 	for i := range len(s) {
// 		s[i] = a[from + i]
// 	}

// 	return
// }

func (a Array) isRhelType() {}
