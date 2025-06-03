package rheltypes

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
)

type Token []byte

func (t Token) Prefix() rhelPrefix {
	for _, p := range rhelPrefixIndex {
		if bytes.HasPrefix(t, []byte(p)) {
			return p
		}
	}
	return rhelPrefix(t)
}

func (t Token) CutPrefix(prefix rhelPrefix) (token Token, ok bool) {
	return bytes.CutPrefix(t, []byte(prefix))
}

func (t Token) ReadSize(prefix rhelPrefix) (size int, err error) {
	sizeBytes, found := t.CutPrefix(prefix)
	if !found {
		return 0, NewPrefixError(prefix, rhelPrefix(sizeBytes))
	}

	size, err = strconv.Atoi(string(sizeBytes))
	if err != nil {
		return 0, fmt.Errorf(
			"failed to convert %q into size: %w",
			sizeBytes,
			err,
		)
	}

	return
}

type TokenIterator struct {
	content [][]byte
	index   int
}

func NewTokenIterator(content []byte) (*TokenIterator, error) {
	data, found := bytes.CutSuffix(content, rhelFieldSep)
	if !found {
		return nil, fmt.Errorf("expected % x terminator", rhelFieldSep)
	}
	items := bytes.Split(data, rhelFieldSep)

	if len(items) == 0 {
		return nil, fmt.Errorf("expected at least 1 item")
	}

	return &TokenIterator{
		content: items,
		index:   0,
	}, nil
}

func (i TokenIterator) Left() int {
	return len(i.content) - i.index
}

func (i TokenIterator) HasNext(size int) bool {
	return len(i.content)-i.index > 0
}

func (i *TokenIterator) Skip(size int) bool {
	if !i.HasNext(size) {
		return false
	}

	i.index += size
	return true
}

func (i TokenIterator) Current() Token {
	return i.content[i.index]
}

func (i *TokenIterator) Read(size int) (tokens []Token, ok bool) {
	if ok = i.Skip(2); !ok {
		return
	}
	data := (i.content[i.index-size : i.index])

	tokens = make([]Token, len(data))

	for i, d := range data {
		tokens[i] = Token(d)
	}

	return
}

func (i *TokenIterator) Next() (token Token, ok bool) {
	if ok = i.Skip(1); !ok {
		return
	}
	token = Token(i.content[i.index-1])

	return
}

func (i *TokenIterator) NextSize(prefix rhelPrefix) (size int, err error) {
	sizeToken, ok := i.Next()
	if !ok {
		return 0, fmt.Errorf("failed to read size token")
	}
	return sizeToken.ReadSize(prefix)
}

func (i TokenIterator) Content() []byte {
	return bytes.Join(i.content, rhelFieldSep)
}

func (i TokenIterator) Dump() string {
	data := i.Content()
	return fmt.Sprintf(
		"index: %d | %q\n%s",
		i.index,
		string(data),
		hex.Dump(data),
	)
}
