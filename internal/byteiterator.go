package internal

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type ByteIterator struct {
	buf    *bufio.Reader
	Offset int
}

func NewByteIteratorFromBytes(data []byte) (iter *ByteIterator) {
	iter = &ByteIterator{
		buf:    bufio.NewReader(bytes.NewReader(data)),
		Offset: 0,
	}

	return
}

func NewByteIteratorFromFile(data *os.File) *ByteIterator {
	return &ByteIterator{buf: bufio.NewReader(data), Offset: 0}
}

func (r *ByteIterator) readBytes(n int) ([]byte, error) {
	buf := make([]byte, n)

	b, err := r.buf.Read(buf)
	if err != nil {
		return nil, err
	} else if b < n {
		return nil, fmt.Errorf("expected to read %d bytes, got %d", n, b)
	}

	log.Printf("read %d bytes, offset %d", n, r.Offset)

	r.Offset += b

	return buf, err
}

func (r *ByteIterator) readByte() (b byte, err error) {
	b, err = r.buf.ReadByte()
	log.Printf("read 1 byte, offset %d: %08b %X", r.Offset, b, b)

	r.Offset++

	return
}

type RdbValueType int

const (
	StringEncoding RdbValueType = iota
	ListEncoding
	SetEncoding
	SortedSetEncoding
	HashEncoding
	_
	_
	_
	_
	ZipmapEncoding
	ZiplistEncoding
	IntsetEncoding
	SortedSetInZiplistEncoding
	HashmapInZiplistEncoding
	ListInQuicklistEncoding
	MetadataEncoding              = 0xFA
	SizesSectionEncoding          = 0xFB
	ExpirationMiliSectionEncoding = 0xFC
	ExpirationSectionEncoding     = 0xFD
	DatabaseSectionEncoding       = 0xFE
	EndOfFileEncoding             = 0xFF
)

const (
	indicatorSize6Bit     = 0b00
	indicatorSize14Bit    = 0b01
	indicatorSize4Bytes   = 0b10
	indicatorSizeNumber   = 0b11
	indicatorSizeInt8Bit  = 0xC0
	indicatorSizeInt16Bit = 0xC1
	indicatorSizeInt32Bit = 0xC2
	indicatorSizeL2F      = 0xC3
	sizeInt8Bit           = 1
	sizeInt16Bit          = 2
	sizeInt32Bit          = 4
	sizeInt64Bit          = 8
	maskFirstTwoBits      = 0x3F
)

type RdbSizeEncoding int

const (
	StringSizeEncoded RdbSizeEncoding = iota
	IntegerSizeEncoded
)

type RdbValueSize struct {
	size     int
	encoding RdbSizeEncoding
}

func (r *ByteIterator) readIntegerSize(b byte) (size RdbValueSize, err error) {
	size.encoding = IntegerSizeEncoded

	switch b {
	case indicatorSizeInt8Bit:
		size.size = sizeInt8Bit
	case indicatorSizeInt16Bit:
		size.size = sizeInt16Bit
	case indicatorSizeInt32Bit:
		size.size = sizeInt32Bit
	case indicatorSizeL2F:
		err = fmt.Errorf("L2F size encoding not supported")
	default:
		err = fmt.Errorf(
			"unknown size encoding: %08b %X",
			b,
			b,
		)
	}

	return
}

func (r *ByteIterator) readSize() (size RdbValueSize, err error) {
	sizeByte, err := r.readByte()
	if err != nil {
		return size, fmt.Errorf("failed to read size byte: %w", err)
	}

	log.Printf("sizeByte +%d %08b %X", r.Offset-1, sizeByte, sizeByte)

	switch sizeByte >> 6 {
	case indicatorSize6Bit:
		size.size = int(sizeByte & maskFirstTwoBits)
	case indicatorSizeNumber:
		size, err = r.readIntegerSize(sizeByte)
	case indicatorSize14Bit:
		sizeByte &= maskFirstTwoBits

		if b, err := r.readByte(); err != nil {
			return size, fmt.Errorf(
				"failed to read 14 bit string size: %w",
				err,
			)
		} else {
			size.size = int(binary.BigEndian.Uint16([]byte{b, sizeByte}))
		}
	case indicatorSize4Bytes:
		if b, err := r.readBytes(sizeInt32Bit); err != nil {
			return size, fmt.Errorf(
				"failed to read 4 bytes string size: %w",
				err,
			)
		} else {
			size.size = int(binary.BigEndian.Uint32(b))
		}
	}

	return size, err
}

type RdbValue interface {
	isRbdValue()
	String() string
}

type RdbStringValue string

func (v RdbStringValue) String() string {
	return string(v)
}

func (v RdbStringValue) isRbdValue() {}

func newRdbIntegerValue(buf []byte) (value RdbStringValue, err error) {
	switch bufSize := len(buf); bufSize {
	case sizeInt8Bit:
		value = RdbStringValue(strconv.Itoa(int(buf[0])))
	case sizeInt16Bit:
		value = RdbStringValue(strconv.Itoa(int(binary.BigEndian.Uint16(buf))))
	case sizeInt32Bit:
		value = RdbStringValue(strconv.Itoa(int(binary.BigEndian.Uint32(buf))))
	default:
		err = fmt.Errorf(
			"failed to create rdb integer string: byte size %d is incorrect",
			bufSize,
		)
	}

	return
}

func (r *ByteIterator) readStringValue() (value RdbStringValue, err error) {
	valueSize, err := r.readSize()
	if err != nil {
		return value, fmt.Errorf("failed to read value size: %w", err)
	}

	buf, err := r.readBytes(valueSize.size)
	if err != nil {
		return value, fmt.Errorf("failed to read value bytes: %w", err)
	}

	switch valueSize.encoding {
	case StringSizeEncoded:
		value = RdbStringValue(buf)
	case IntegerSizeEncoded:
		value, err = newRdbIntegerValue(buf)
	}

	return
}

type RdbExpirationTime int64

func (v RdbExpirationTime) String() string {
	return strconv.Itoa(int(v))
}

func (v RdbExpirationTime) isRbdValue() {}

func newRdbExpirationTime(
	iter *ByteIterator,
) (value RdbExpirationTime, err error) {
	expBytes, err := iter.readBytes(sizeInt32Bit)
	if err != nil {
		return value, err
	}

	exp := binary.LittleEndian.Uint32(expBytes)

	value = RdbExpirationTime(time.Duration(exp) * time.Millisecond)

	return
}

func newRdbExpirationTimeMili(
	iter *ByteIterator,
) (value RdbExpirationTime, err error) {
	expBytes, err := iter.readBytes(sizeInt64Bit)
	if err != nil {
		return value, err
	}

	exp := binary.LittleEndian.Uint64(expBytes)

	if exp > 1<<63-1 {
		return value, fmt.Errorf("expiry timestamp too large: %d", exp)
	}

	value = RdbExpirationTime(exp)

	return
}

func (r *ByteIterator) readKeyValue() (key RdbStringValue, value RdbValue, err error) {
	encodingByte, err := r.readByte()
	if err != nil {
		return key, value, fmt.Errorf("failed to read value encoding: %w", err)
	}

	encoding := RdbValueType(encodingByte)
	if encoding == EndOfFileEncoding || encoding == DatabaseSectionEncoding {
		return key, value, ErrEndOfSection
	}

	switch encoding {
	case StringEncoding, MetadataEncoding:
		if key, err = r.readStringValue(); err != nil {
			return key, value, fmt.Errorf("failed to read key: %w", err)
		}

		value, err = r.readStringValue()
	case ExpirationMiliSectionEncoding:
		value, err = newRdbExpirationTimeMili(r)
	case ExpirationSectionEncoding:
		value, err = newRdbExpirationTime(r)
	default:
		err = fmt.Errorf(
			"encoding %08b %X not implemented",
			encodingByte,
			encodingByte,
		)
	}

	return key, value, err
}
