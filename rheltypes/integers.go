package rheltypes

import (
	"fmt"
	"strconv"
)

type Integer int

func NewIntegerFromTokens(token Token) (i Integer, err error) {
	val, err := token.AsSize()
	if err != nil {
		return i, fmt.Errorf("failed to convert to integer %q: %w", token, err)
	}

	i = Integer(val)

	return
}

func (i Integer) Size() int {
	sizeStr := strconv.Itoa(int(i))

	return len(
		IntegerPrefix,
	) + len(
		sizeStr,
	) + len(
		rhelFieldDelim,
	)
}

func (i Integer) Serialize() []byte {
	buf := make([]byte, 0, i.Size())

	return fmt.Appendf(buf, "%s%d\r\n", IntegerPrefix, i)
}

func (i Integer) String() string {
	return strconv.Itoa(int(i))
}

func (i Integer) First() RhelType {
	return i
}

func (i Integer) Integer() (int, error) {
	return int(i), nil
}

func (Integer) isRhelType() {}
