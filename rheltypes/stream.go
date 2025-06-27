package rheltypes

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	streamItemSliceSize   = 2
	streamArrayItemSize   = 2
	defaultStreamCapacity = 256
	defaultIdSep          = "-"
)

var (
	wrongIdError = fmt.Errorf(
		"The ID specified in XADD is equal or smaller than the target stream top item",
	)

	zeroIdError = fmt.Errorf(
		"The ID specified in XADD must be greater than 0-0",
	)
)

type IdGeneration int

const (
	ExplicitId IdGeneration = iota
	ZeroId
	BlankId
	BlankTsId
	BeginId
	EndId
)

type StreamItemId struct {
	ts  int
	seq int
}

func NewStreamItemId(query string) (id StreamItemId, idType IdGeneration) {
	switch query {
	case "0-0":
		idType = ZeroId
	case "*":
		idType = BlankId
	case "-":
		idType = ZeroId
	case "+":
		idType = EndId
		id.ts = -1
		id.seq = -1
	default:
		idType = ExplicitId
	}

	if idType != ExplicitId {
		return id, idType
	}

	ts, seq, _ := strings.Cut(query, defaultIdSep)
	id.ts, _ = strconv.Atoi(ts)
	id.seq = -1

	switch seq {
	case "*", "":
		idType = BlankTsId
	default:
		id.seq, _ = strconv.Atoi(seq)
	}

	return id, idType
}

func (id StreamItemId) IsZero() bool {
	return id.ts+id.seq == 0
}

func (id StreamItemId) LessTs(other StreamItemId) bool {
	return id.ts < other.ts
}

func (id StreamItemId) LessSeq(other StreamItemId) bool {
	return id.ts == other.ts && id.seq < other.seq
}

func (id StreamItemId) Less(other StreamItemId) bool {
	return id.LessTs(other) || id.LessSeq(other)
}

func (id StreamItemId) NextSeq() StreamItemId {
	return StreamItemId{id.ts, id.seq + 1}
}

func (id StreamItemId) Next() StreamItemId {
	return StreamItemId{int(time.Now().UnixMilli()), 0}
}

func (id StreamItemId) ToString() string {
	return strconv.Itoa(id.ts) + defaultIdSep + strconv.Itoa(id.seq)
}

func (id StreamItemId) Cmp(other StreamItemId) int {
	if other.ts == -1 {
		return -1
	}

	if id.ts != other.ts {
		return cmp.Compare(id.ts, other.ts)
	}

	if other.seq == -1 {
		return -1
	}

	return cmp.Compare(id.seq, other.seq)
}

type StreamItem struct {
	id     StreamItemId
	values map[string]string
}

// func NewStreamItemFromArray(id string, items Array) (item StreamItem) {
// 	item.id = NewStreamItemId(id)
// 	item.values = make(map[string]string, len(items)/streamItemSliceSize)

// 	for pair := range slices.Chunk(items, streamItemSliceSize) {
// 		item.values[pair[0].String()] = pair[1].String()
// 	}

// 	return
// }

func (i StreamItem) Size() int {
	return len(i.values)
}

// func (i StreamItem) Serialize() []byte {
// 	return nil
// }

// func (i StreamItem) String() string {
// 	return ""
// }

func (i StreamItem) ToArray() (a Array) {
	a = make(Array, streamArrayItemSize)

	a[0] = NewBulkString(i.id.ToString())

	valuesArray := make(Array, 0, len(i.values))

	for key, value := range i.values {
		valuesArray = append(
			valuesArray,
			NewBulkString(key),
			NewBulkString(value),
		)
	}

	a[1] = valuesArray

	return
}

type Stream []StreamItem

func NewStream() Stream {
	return make(Stream, 0, defaultStreamCapacity)
}

func (s Stream) LastId() (id StreamItemId) {
	if len(s) == 0 {
		return StreamItemId{}
	}

	return s.At(-1).id
}

func (s *Stream) GenerateId(query string) (id StreamItemId, err error) {
	var genType IdGeneration
	id, genType = NewStreamItemId(query)
	lastId := s.LastId()

	switch genType {
	case ZeroId:
		return id, zeroIdError
	case ExplicitId:
		if !lastId.Less(id) {
			return id, wrongIdError
		}
	case BlankTsId:
		if lastId.ts == id.ts {
			id = lastId.NextSeq()
		} else if lastId.ts < id.ts {
			id = id.NextSeq()
		} else {
			return id, wrongIdError
		}
	case BlankId:
		id = lastId.Next()
	}

	return id, err
}

func (s *Stream) Add(
	idStr string,
	values map[string]string,
) (added string, err error) {
	id, err := s.GenerateId(idStr)
	if err != nil {
		return
	}

	*s = append(*s, StreamItem{id: id, values: values})

	added = id.ToString()

	return
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

func helperItemIdCompare(item StreamItem, id StreamItemId) int {
	return item.id.Cmp(id)
}

func (s Stream) Range(lower, upper string, includeLower bool) Stream {
	lowerId, lowerIdType := NewStreamItemId(lower)
	upperId, upperIdType := NewStreamItemId(upper)

	lowerId.seq = max(lowerId.seq, 0)

	lowerIndex := 0
	found := false

	if lowerIdType != ZeroId {
		lowerIndex, found = slices.BinarySearchFunc(
			s,
			lowerId,
			helperItemIdCompare,
		)

		if (!found && lowerIndex < len(s)-1) || (found && !includeLower) {
			lowerIndex++
		}
	}

	upperIndex := len(s)

	if upperIdType != EndId {
		upperIndex, found = slices.BinarySearchFunc(
			s,
			upperId,
			helperItemIdCompare,
		)
		if found {
			upperIndex++
		}
	}

	return s[lowerIndex:upperIndex]
}

func (s Stream) ToArray() (a Array) {
	a = make(Array, len(s))

	for i, item := range s {
		a[i] = item.ToArray()
	}

	return
}

func (s Stream) isRhelType() {}
