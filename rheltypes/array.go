package rheltypes

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Array []RhelType

func NewArray(tokens *TokenIterator) (Array, error) {
	log.Printf("Array %s\n", tokens.Dump())

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
	log.Printf("Array constructed: %v\n", output)

	return output, nil
}

func (a Array) isRhelType() {}

func (a Array) Size() int {
	size := 0
	for _, i := range a {
		size += i.Size()
	}
	sizeStr := len(strconv.Itoa(len(a)))

	return len(ArrayPrefix) + sizeStr + len(rhelFieldSep) + size
}

func (a Array) Serialize() []byte {
	buf := make([]byte, 0)

	buf = fmt.Append(
		buf,
		ArrayPrefix,
		[]byte(strconv.Itoa(len(a))),
		rhelFieldSep,
	)

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
