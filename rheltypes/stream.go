package rheltypes

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	streamItemSliceSize   = 2
	defaultStreamCapacity = 256
)

var (
	wrongIdError = fmt.Errorf(
		"The ID specified in XADD is equal or smaller than the target stream top item",
	)

	ZeroIdError = fmt.Errorf(
		"The ID specified in XADD must be greater than 0-0",
	)
)

type IdGeneration int

const (
	GenerateNext IdGeneration = iota
	GenerateNextSeq
	Explicit
)

type StreamItemId struct {
	ts  int
	seq int
}

func NewStreamItemId(query string) (id StreamItemId, idType IdGeneration) {
	idType = Explicit
	if query == "*" {
		idType = GenerateNext
		id.seq = 1

		return
	}

	ts, seq, _ := strings.Cut(query, "-")
	id.ts, _ = strconv.Atoi(ts)

	if seq == "*" {
		idType = GenerateNextSeq
	} else {
		id.seq, _ = strconv.Atoi(seq)
	}

	return
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

func (s Stream) LastId() (id StreamItemId) {
	if len(s) == 0 {
		return StreamItemId{}
	}

	return s.At(-1).id
}

func (s *Stream) GenerateId(query string) (id StreamItemId, err error) {
	if query == "0-0" {
		return id, ZeroIdError
	}

	var genType IdGeneration
	id, genType = NewStreamItemId(query)
	lastId := s.LastId()

	log.Println(id, genType, lastId)

	switch genType {
	case Explicit:
		if !lastId.Less(id) {
			return id, wrongIdError
		}
	case GenerateNextSeq:
		if !lastId.LessTs(id) {
			return id, wrongIdError
		}

		id = lastId.NextSeq()
	case GenerateNext:
		id = lastId.Next()
	}

	return
}

func (s *Stream) Add(idStr string, values map[string]string) (err error) {
	id, err := s.GenerateId(idStr)
	if err != nil {
		return err
	}

	*s = append(*s, StreamItem{id: id, values: values})

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

func (s Stream) isRhelType() {}
