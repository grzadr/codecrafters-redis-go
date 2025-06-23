package rheltypes

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
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

const defaultIteratorBufferSize = 256

type BuffIterator struct {
	buf    *bufio.Reader
	done   bool
	offset int
}

func NewBuffIterator(data []byte) (iter *BuffIterator) {
	iter = &BuffIterator{
		buf:  bufio.NewReader(bytes.NewReader(data)),
		done: false,
	}

	return
}

func (r *BuffIterator) Offset() int {
	return r.offset
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

	r.offset += rn

	return
}

func identicalSlices(a, b []byte) bool {
	return slices.Compare(a, b) == 0
}

func (r *BuffIterator) readString(delim []byte) (out string, err error) {
	if r.IsDone() {
		return
	}

	buf := make([]byte, 0, defaultIteratorBufferSize)
	delimLen := len(delim)
	d := delim[delimLen-1]

	for {
		temp, err := r.buf.ReadBytes(d)
		r.offset += len(temp)
		buf = append(buf, temp...)

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

	if ok = identicalSlices(b, delim); !ok {
		return
	}

	n, err := r.buf.Discard(len(delim))

	r.offset += n

	return
}

type TokenIterator struct {
	*BuffIterator
}

func NewTokenIterator(content []byte) *TokenIterator {
	return &TokenIterator{BuffIterator: NewBuffIterator(content)}
}

func (i *TokenIterator) NextToken() (token Token, err error) {
	str, err := i.readString(rhelFieldDelim)
	if err != nil || i.IsDone() {
		return
	}

	token = NewToken(str)

	return
}
