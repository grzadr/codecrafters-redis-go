package rheltypes

import (
	"fmt"
	"strconv"
)

type Integer int

func NewIntegerFromTokens(tokens *TokenIterator) (i Integer, err error) {
	wrap := func(err error) error {
		if err != nil {
			return fmt.Errorf("failed to create integer: %w", err)
		}

		return nil
	}
	numToken, ok := tokens.Next()

	if !ok {
		return i, wrap(fmt.Errorf("failed to load next token"))
	}

	numToken, ok = numToken.CutPrefix(IntegerPrefix)

	if !ok {
		return i, wrap(NewPrefixError(IntegerPrefix, rhelPrefix(numToken)))
	}

	value, err := strconv.Atoi(string(numToken))
	if err != nil {
		return i, wrap(
			fmt.Errorf("failed to convert integer %q: %w", numToken, err),
		)
	}

	i = Integer(value)

	return
}

func (i Integer) Size() int {
	sizeStr := strconv.Itoa(int(i))

	return len(
		IntegerPrefix,
	) + len(
		sizeStr,
	) + len(
		rhelFieldSep,
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
