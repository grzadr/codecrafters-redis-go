package rheltypes

import "fmt"

type Array []RhelType

func NewArray(content []byte) (Array, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("cannot create array from empty slice")
	}

	if content[0] != '*' {
		return nil, fmt.Errorf(
			"array must start with '*' not '%s",
			string(content[0]),
		)
	}

	return nil, nil
}

func (a Array) isRhelType() {}

func (a Array) Serialize() []byte {
	return nil
}
