package rheltypes

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"slices"
	"strconv"
	"strings"
)

type Token struct {
	Prefix rhelPrefix
	Data   string
}

func NewToken(str string) (token Token) {
	before, _ := strings.CutSuffix(str, string(rhelFieldDelim))
	token.Prefix = NewRhelPrefix(before[:1])
	token.Data = before[1:]

	return
}

func (t Token) ToString() string {
	return fmt.Sprintf("%s%s", t.Prefix, t.Data)
}

func (t Token) AsSize() (i int, err error) {
	i, err = strconv.Atoi(t.Data)
	if err != nil {
		err = fmt.Errorf("failed to convert token %s to size: %w", t, err)
	}

	return
}

// func (t Token) ReadSize(prefix rhelPrefix) (size int, err error) {
// 	sizeBytes, found := t.CutPrefix(prefix)
// 	if !found {
// 		return 0, NewPrefixError(prefix, rhelPrefix(sizeBytes))
// 	}

// 	size, err = strconv.Atoi(string(sizeBytes))
// 	if err != nil {
// 		return 0, fmt.Errorf(
// 			"failed to convert %q into size: %w",
// 			sizeBytes,
// 			err,
// 		)
// 	}

// 	return
// }

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

func (r *BuffIterator) validate(err error) error {
	switch err {
	case io.EOF:
		r.done = true
	case nil:
	default:
		return err
	}

	return nil
}

func (r *BuffIterator) readBytes(n int) (b []byte, err error) {
	if r.IsDone() {
		return
	}

	b = make([]byte, n)

	rn, err := r.buf.Read(b)
	if err := r.validate(err); err != nil {
		return nil, fmt.Errorf("failed to read %d bytes: %w", n, err)
	}

	b = b[:rn]

	return
}

// func (r *BuffIterator) readByte() (b byte, err error) {
// 	if r.IsDone() {
// 		return
// 	}

// 	b, err = r.buf.ReadByte()

// 	if err := r.validate(err); err != nil {
// 		return 0, fmt.Errorf("failed to read byte: %w", err)
// 	}

// 	return
// }

func identicalSlices(a, b []byte) bool {
	// return slices.Compare(buf[len(buf)-delimLen:], delim) == 0
	return slices.Compare(a, b) == 0
}

func (r *BuffIterator) readString(delim []byte) (out string, err error) {
	if r.IsDone() {
		log.Println("buffer is done")

		return
	}

	buf := make([]byte, 0, defaultIteratorBufferSize)
	delimLen := len(delim)
	d := delim[delimLen-1]

	for {
		temp, err := r.buf.ReadBytes(d)
		buf = append(buf, temp...)

		// log.Println("buf: ")

		if err = r.validate(err); err != nil || r.IsDone() {
			break
		} else if identicalSlices(buf[len(buf)-delimLen:], delim) {
			break
		}
	}

	out = string(buf)

	return
}

func (r *BuffIterator) skipDelim(delim []byte) (ok bool, err error) {
	b, err := r.buf.Peek(len(delim))

	if err = r.validate(err); err != nil {
		return ok, fmt.Errorf(
			"failed to skip delim %q %X: %w",
			delim,
			delim,
			err,
		)
	}

	return identicalSlices(b, delim), nil
}

type TokenIterator struct {
	// content [][]byte
	// index   int
	*BuffIterator
	// LastToken Token
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

// func (i *TokenIterator) Read(size int) (tokens []Token, ok bool) {
// 	// if ok = i.Skip(size); !ok {
// 	// 	return
// 	// }

// 	data := (i.content[i.index-size : i.index])

// 	tokens = make([]Token, len(data))

// 	for i, d := range data {
// 		tokens[i] = Token(d)
// 	}

// 	return
// }

func (i *TokenIterator) NextToken() (token Token, err error) {
	str, err := i.readString(rhelFieldDelim)
	if err != nil {
		return
	}

	log.Printf("token: %q\n", str)

	token = NewToken(str)
	// i.LastToken = token
	// if ok = i.Skip(1); !ok {
	// 	return
	// }
	// token = Token(i.content[i.index-1])
	return
}

// func (i *TokenIterator) NextSize(prefix rhelPrefix) (size int, err error) {
// 	sizeToken, ok := i.Next()
// 	if !ok {
// 		return 0, fmt.Errorf("failed to read size token")
// 	}

// 	return sizeToken.ReadSize(prefix)
// }

// func (i TokenIterator) Content() []byte {
// 	return bytes.Join(i.content, rhelFieldSep)
// }

// func (i TokenIterator) Dump() string {
// 	data := i.Content()

// 	return fmt.Sprintf(
// 		"index: %d | %q\n%s",
// 		i.index,
// 		string(data),
// 		hex.Dump(data),
// 	)
// }
