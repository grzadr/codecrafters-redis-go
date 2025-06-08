package internal

import "fmt"

type RdbHeader struct {
	magic   string
	version string
}

const (
	DEFAULT_HEADER_MAGIC   = "REDIS"
	HEADER_MAGIC_BYTES     = 5
	DEFAULT_HEADER_VERSION = "0011"
	HEADER_VERSION_BYTES   = 4
	HEADER_BYTES           = HEADER_MAGIC_BYTES + HEADER_VERSION_BYTES
	METADATA_FIELD         = 0xFA
)

func NewRdbHeader() RdbHeader {
	return RdbHeader{
		magic:   DEFAULT_HEADER_MAGIC,
		version: DEFAULT_HEADER_VERSION,
	}
}

func ReadRdbHeader(iter *ByteIterator) (h RdbHeader, err error) {
	data, err := iter.readBytes(HEADER_MAGIC_BYTES)
	if err != nil {
		return
	}

	h.magic = string(data)

	data, err = iter.readBytes(HEADER_VERSION_BYTES)
	if err != nil {
		return
	}

	h.version = string(data)

	return
}

type RdbMetadata map[string]string

func readRdbMetadata(iter *ByteIterator) (meta RdbMetadata, err error) {
	if meta_field, err := iter.readByte(); err != nil {
		return nil, err
	} else if meta_field != METADATA_FIELD {
		return nil, fmt.Errorf("expected field %X, got %X", METADATA_FIELD, meta_field)
	}

	return
}

type RdbFile struct {
	header   RdbHeader
	metadata RdbMetadata
	selector int
}

func ReadRdbFile(iter *ByteIterator) (rdb *RdbFile, err error) {
	rdb = &RdbFile{}

	if rdb.header, err = ReadRdbHeader(iter); err != nil {
		return nil, fmt.Errorf("error reading rdb header: %w", err)
	}

	if rdb.metadata, err = readRdbMetadata(iter); err != nil {
		return nil, fmt.Errorf("error reading rdb metadata: %w", err)
	}

	return
}
