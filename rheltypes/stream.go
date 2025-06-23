package rheltypes

import "slices"

type StreamItem struct {
	id     string
	values map[string]string
}

const (
	streamItemSliceSize   = 2
	defaultStreamCapacity = 256
)

func NewStreamItemFromArray(id string, items Array) (item StreamItem) {
	item.id = id
	item.values = make(map[string]string, len(items)/streamItemSliceSize)

	for pair := range slices.Chunk(items, streamItemSliceSize) {
		item.values[pair[0].String()] = pair[1].String()
	}

	return
}

func (i StreamItem) Size() int {
	return 0
}

func (i StreamItem) Serialize() []byte {
	return nil
}

func (i StreamItem) String() string {
	return ""
}

type Stream []StreamItem

func NewStream() Stream {
	return make(Stream, 0, defaultStreamCapacity)
}

func (s Stream) Size() int {
	return 0
}

func (s Stream) Serialize() []byte {
	return nil
}

func (s Stream) String() string {
	return ""
}

func (s Stream) At(index int) *StreamItem {
	if index < 0 {
		return s.At(len(s) + index)
	} else if index < len(s) {
		return &s[index]
	} else {
		return nil
	}
}

func (s Stream) First() RhelType {
	return nil
}

func (s Stream) TypeName() string {
	return "stream"
}

func (s Stream) Integer() (int, error) {
	return 0, nil
}

func (s Stream) isRhelType() {}
