package internal

import (
	"errors"
	"fmt"
	"iter"
	"log"
	"maps"
)

type RdbHeader struct {
	magic   string
	version string
}

const (
	HEADER_MAGIC_CONTENT   = "REDIS"
	HEADER_MAGIC_SIZE      = len(HEADER_MAGIC_CONTENT)
	HEADER_VERSION_CONTENT = "0011"
	HEADER_VERSION_SIZE    = len(HEADER_VERSION_CONTENT)
	DEFAULT_METADATA_SIZE  = 8
	CHECKSUM_SIZE          = 8
	METADATA_SECTION       = 0xFA
	SIZES_SECTION          = 0xFB
	EXP_MILI_SECTION       = 0xFC
	EXP_SECTION            = 0xFD
	DATABASE_SECTION       = 0xFE
	EOF_SECTION            = 0xFF
)

func NewRdbHeader() RdbHeader {
	return RdbHeader{
		magic:   HEADER_MAGIC_CONTENT,
		version: HEADER_VERSION_CONTENT,
	}
}

func ReadRdbHeader(iter *ByteIterator) (h RdbHeader, err error) {
	data, err := iter.readBytes(HEADER_MAGIC_SIZE)
	if err != nil {
		return
	}

	h.magic = string(data)

	if h.magic != HEADER_MAGIC_CONTENT {
		return h, fmt.Errorf(
			"expected %s magic string, got %s",
			h.magic,
			HEADER_MAGIC_CONTENT,
		)
	}

	data, err = iter.readBytes(HEADER_VERSION_SIZE)
	if err != nil {
		return
	}

	h.version = string(data)

	return
}

type RdbMetadata map[string]string

func readRdbMetadata(iter *ByteIterator) (meta RdbMetadata, err error) {
	meta = make(RdbMetadata, DEFAULT_METADATA_SIZE)

	for {
		key, value, err := iter.readKeyValue()
		if errors.Is(err, ErrEndOfSection) {
			break
		} else if err != nil {
			return nil, err
		}

		meta[key.String()] = value.String()

		log.Printf("added metadata %q - %q\n", key.String(), value.String())
	}

	log.Println("metadata completed")

	return
}

var ErrEndOfSection = errors.New("end of database")

type RdbKeyValue struct {
	Expiry int64
	Value  string
}

func newRdbKeyValue(
	iter *ByteIterator,
) (key string, value RdbKeyValue, err error) {
	for {
		rawKey, rawValue, err := iter.readKeyValue()
		if err != nil {
			return key, value, err
		}

		switch v := rawValue.(type) {
		case RdbExpirationTime:
			value.Expiry = int64(v)
		case RdbStringValue:
			value.Value = rawValue.String()
			key = rawKey.String()

			return key, value, err
		}
	}
}

type RdbFileValues map[string]RdbKeyValue

func newRdbFileValues(
	iter *ByteIterator,
	size int,
) (values RdbFileValues, err error) {
	values = make(RdbFileValues, size)

	for {
		k, v, err := newRdbKeyValue(iter)
		if errors.Is(err, ErrEndOfSection) {
			break
		}

		if err != nil {
			return nil, err
		}

		values[k] = v
	}

	return
}

type RdbFile struct {
	header      RdbHeader
	metadata    RdbMetadata
	selector    int
	hashSize    int
	expKeysSize int
	values      RdbFileValues
	checksum    [CHECKSUM_SIZE]byte
}

func (f *RdbFile) setSelector(iter *ByteIterator) (err error) {
	log.Println("reading selector size byte")

	size, err := iter.readSize()
	if err != nil {
		return err
	}

	f.selector = int(size.size)

	return
}

func (f *RdbFile) setSizes(iter *ByteIterator) (err error) {
	markByte, err := iter.readByte()
	if err != nil {
		return err
	}

	if markByte != SIZES_SECTION {
		return fmt.Errorf(
			"expected %X mark, got %08b %X",
			SIZES_SECTION,
			markByte,
			markByte,
		)
	}

	items := []struct {
		ptr  *int
		name string
	}{
		{ptr: &f.hashSize, name: "hash"},
		{ptr: &f.expKeysSize, name: "exp keys"},
	}

	for _, field := range items {
		size, err := iter.readSize()
		if err != nil {
			return fmt.Errorf("failed to read %s size: %w", field.name, err)
		}

		*field.ptr = int(size.size)
	}

	log.Println("sizes completed")

	return err
}

func (f *RdbFile) setChecksum(iter *ByteIterator) (err error) {
	checkBytes, err := iter.readBytes(CHECKSUM_SIZE)
	if err != nil {
		return err
	}

	f.checksum = [8]byte(checkBytes)

	return
}

func (f RdbFile) Iter() iter.Seq2[string, RdbKeyValue] {
	return maps.All(f.values)
}

func ReadRdbFile(iter *ByteIterator) (rdb *RdbFile, err error) {
	rdb = &RdbFile{}

	if rdb.header, err = ReadRdbHeader(iter); err != nil {
		return nil, fmt.Errorf("error reading rdb header: %w", err)
	}

	if rdb.metadata, err = readRdbMetadata(iter); err != nil {
		return nil, fmt.Errorf("error reading rdb metadata: %w", err)
	}

	if err = rdb.setSelector(iter); err != nil {
		return nil, fmt.Errorf("error reading rdb selector: %w", err)
	}

	if err = rdb.setSizes(iter); err != nil {
		return nil, fmt.Errorf("error reading rdb size info: %w", err)
	}

	if rdb.values, err = newRdbFileValues(iter, rdb.hashSize); err != nil {
		return nil, fmt.Errorf("error reading rdb key values: %w", err)
	}

	if err = rdb.setChecksum(iter); err != nil {
		return nil, fmt.Errorf("error reading rdb checksum: %w", err)
	}

	return
}
