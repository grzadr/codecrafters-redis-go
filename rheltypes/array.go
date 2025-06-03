package rheltypes

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Array []RhelType

func NewArray(content []byte) (Array, error) {
	log.Printf("Array %q:\n%s\n", content, hex.Dump(content))
	data, err := cutRhelPrefix(content, ArrayPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to create array: %w", err)
	}

	sizeBytes, elements, found := bytes.Cut(data, rhelFieldSep)

	if !found {
		return nil, NewContentError(content)
	}

	size, err := strconv.Atoi(string(sizeBytes))
	if err != nil {
		return nil, fmt.Errorf(
			"failed to convert %s into number",
			string(sizeBytes),
		)
	}

	output := make(Array, 0, size)
	var itemBytes []byte
	var item RhelType
	for range size {
		itemBytes, elements, found = bytes.Cut(elements, rhelFieldSep)
		if !found {
			return nil, fmt.Errorf(
				"missing expected field separator in %s (% x)",
				string(content),
				content,
			)
		}
		if item, err = RhelSerialize(content); err != nil {
			return nil, fmt.Errorf(
				"failed to serialize %s (% x)",
				string(itemBytes),
				itemBytes,
			)
		}
		output = append(output, item)
	}

	if len(elements) > 0 {
		return nil, fmt.Errorf(
			"unprocessed content left from %s: %s (% x)",
			string(content),
			string(elements),
			elements,
		)
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
