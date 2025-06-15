package rheltypes

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"slices"
	"strconv"
)

type Token struct {
	Prefix rhelPrefix
	Data   []byte
}

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

func NewToken() (token Token, err error) {
}

const defaultIteratorBufferSize = 256

type BuffIterator struct {
	buf  *bufio.Reader
	done bool
}

func NewBuffIterator(data []byte) (iter *BuffIterator) {
	iter = &BuffIterator{
		buf:  bufio.NewReader(bytes.NewReader(data)),
		done: false,
	}

	return
}

func (r *BuffIterator) IsDone() bool {
	return r.done
}

func (r *BuffIterator) validate(err error, n int) error {
	switch err {
	case io.EOF:
		r.done = true
	case nil:
	default:
		return fmt.Errorf("failed to read %d bytes: %w", n, err)
	}

	return nil
}

func (r *BuffIterator) readBytes(n int) ([]byte, error) {
	buf := make([]byte, n)

	b, err := r.buf.Read(buf)
	if err := r.validate(err, n); err != nil {
		return nil, err
	}

	return buf[:b], err
}

func (r *BuffIterator) readByte() (b byte, err error) {
	b, err = r.buf.ReadByte()

	if err := r.validate(err, 1); err != nil {
		return 0, err
	}

	return
}

func (r *BuffIterator) readString(delim []byte) (out string, err error) {
	buf := make([]byte, 0, defaultIteratorBufferSize)
	delimLen := len(delim)
	d := delim[delimLen-1]

	for {
		temp, err := r.buf.ReadBytes(d)
		buf = append(buf, temp...)

		if err = r.validate(err, -1); err != nil || r.IsDone() {
			break
		} else if slices.Compare(buf[len(buf)-delimLen:], delim) == 0 {
			break
		}
	}

	return
}

type TokenIterator struct {
	// content [][]byte
	// index   int
	*BuffIterator
}

func NewTokenIterator(content []byte) *TokenIterator {
	// data, found := bytes.CutSuffix(content, rhelFieldSep)
	// if !found {
	// 	return nil, fmt.Errorf("expected % x terminator", rhelFieldSep)
	// }
	// items := bytes.Split(data, rhelFieldSep)
	// if len(items) == 0 {
	// 	return nil, fmt.Errorf("expected at least 1 item")
	// }
	return &TokenIterator{BuffIterator: NewBuffIterator(content)}
}

// func (i TokenIterator) Left() int {
// 	return len(i.content) - i.index
// }

// func (i TokenIterator) HasNext(size int) bool {
// 	return len(i.content)-i.index > 0
// }

// func (i *TokenIterator) Skip(size int) bool {
// 	if !i.HasNext(size) {
// 		return false
// 	}

// 	i.index += size

// 	return true
// }

// func (i TokenIterator) Current() Token {
// 	return i.content[i.index]
// }

func (i *TokenIterator) Read(size int) (tokens []Token, ok bool) {
	if ok = i.Skip(size); !ok {
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
