package rheltypes

import (
	"cmp"
	"log"
	"slices"
	"strconv"
)

const defaultSortedSetCapacity = 16

type SortedSetMember struct {
	name  string
	score float64
}

func (m SortedSetMember) AsBulkString() BulkString {
	return NewBulkString(strconv.FormatFloat(m.score, 'e', 2, 64))
}

func (m SortedSetMember) asArray() Array {
	return Array{
		NewBulkString(m.name),
		m.AsBulkString(),
	}
}

type SortedSet struct {
	names   map[string]*SortedSetMember
	members []*SortedSetMember
}

func NewSortedSet() *SortedSet {
	return &SortedSet{
		names:   make(map[string]*SortedSetMember, defaultSortedSetCapacity),
		members: make([]*SortedSetMember, 0, defaultSortedSetCapacity),
	}
}

func (s SortedSet) Get(name string) (member *SortedSetMember, found bool) {
	member, found = s.names[name]

	return
}

func (s SortedSet) FindInsertIndex(
	other *SortedSetMember,
) (index int) {
	index, _ = slices.BinarySearchFunc(
		s.members,
		other,
		func(member, other *SortedSetMember) int {
			log.Printf(
				"%+v %+v %d %d",
				*member,
				*other,
				cmp.Compare(member.score, other.score),
				cmp.Compare(member.name, other.name),
			)

			if scoreCmp := cmp.Compare(member.score, other.score); scoreCmp == 0 {
				return cmp.Compare(member.name, other.name)
			} else {
				return scoreCmp
			}
		},
	)

	log.Println("index:", index)

	return
}

func (s SortedSet) Index(name string) (index int, found bool) {
	member, found := s.Get(name)

	if !found {
		return
	}

	index = s.indexMember(member)

	log.Println("index", name, index)

	return index, index > -1
}

func (s *SortedSet) InsertMember(member *SortedSetMember) {
	s.members = slices.Insert(
		s.members,
		s.FindInsertIndex(member),
		member,
	)
}

func (s *SortedSet) Add(name string, score float64) (found bool) {
	var member *SortedSetMember
	if member, found = s.Get(name); found {
		member.score = score
		index := s.indexMember(member)
		s.members = slices.Delete(s.members, index, index+1)
	} else {
		member = &SortedSetMember{name: name, score: score}
		s.names[name] = member
	}

	s.InsertMember(member)

	return
}

func (s SortedSet) Size() int {
	return len(s.members)
}

func (s SortedSet) First() RhelType {
	if s.Size() == 0 {
		return nil
	}

	return s.members[0].asArray()
}

func (s SortedSet) Float() (float64, error) {
	return 0.0, nil
}

func (s SortedSet) Integer() (int, error) {
	return 0, nil
}

func (s SortedSet) Serialize() []byte {
	return s.asArray().Serialize()
}

func (s SortedSet) String() string {
	return s.asArray().String()
}

func (s SortedSet) TypeName() string {
	return "sortedset"
}

func (s SortedSet) Range(start, stop int) (out Array) {
	return s.asArray().Range(start, stop)
}

func (s SortedSet) asArray() Array {
	output := make(Array, s.Size())

	for i, m := range s.members {
		output[i] = NewBulkString(m.name)
	}

	return output
}

func (s SortedSet) indexMember(member *SortedSetMember) (index int) {
	return slices.IndexFunc(s.members, func(ref *SortedSetMember) bool {
		return ref == member
	})
}

func (s SortedSet) isRhelType() {}
