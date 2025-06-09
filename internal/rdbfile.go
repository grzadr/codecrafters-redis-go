package internal

import (
	"bufio"
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
	defaultRdbHeaderMagicString    = "REDIS"
	defaultRdbHeaderMagicLength    = len(defaultRdbHeaderMagicString)
	defaultRdbHeaderVersionString  = "0011"
	defaultRdbHeaderVersionLength  = len(defaultRdbHeaderVersionString)
	defaultRdbFileMetadataCapacity = 8
	defaultRdbFileKeyStoreCapacity = 8
	defaultRdbFileChecksumLength   = 8
)

func newRdbHeader() RdbHeader {
	return RdbHeader{
		magic:   defaultRdbHeaderMagicString,
		version: defaultRdbHeaderVersionString,
	}
}

func (h RdbHeader) decode() (content []byte) {
	content = make([]byte, 0, len(h.magic)+len(h.version))
	content = fmt.Append(content, h.magic, h.version)

	return
}

func readRdbHeader(iter *ByteIterator) (h RdbHeader, err error) {
	data, err := iter.readBytes(defaultRdbHeaderMagicLength)
	if err != nil {
		return
	}

	h.magic = string(data)

	if h.magic != defaultRdbHeaderMagicString {
		return h, fmt.Errorf(
			"expected %s magic string, got %s",
			h.magic,
			defaultRdbHeaderMagicString,
		)
	}

	data, err = iter.readBytes(defaultRdbHeaderVersionLength)
	if err != nil {
		return
	}

	h.version = string(data)

	return
}

type RdbMetadata map[string]string

func newRdbMetadata() RdbMetadata {
	return make(RdbMetadata, defaultRdbFileMetadataCapacity)
}

func readRdbMetadata(iter *ByteIterator) (meta RdbMetadata, err error) {
	meta = newRdbMetadata()

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

func readRdbKeyValue(
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

type RdbFileKeyStore map[string]RdbKeyValue

func newRdbKeyStore() RdbFileKeyStore {
	return make(RdbFileKeyStore, defaultRdbFileKeyStoreCapacity)
}

func readRdbFileValues(
	iter *ByteIterator,
	size int,
) (values RdbFileKeyStore, err error) {
	values = make(RdbFileKeyStore, size)

	for {
		k, v, err := readRdbKeyValue(iter)
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
	keyStore    RdbFileKeyStore
	checksum    [defaultRdbFileChecksumLength]byte
}

func NewRdbfFile() *RdbFile {
	return &RdbFile{
		header:      newRdbHeader(),
		metadata:    newRdbMetadata(),
		selector:    0,
		hashSize:    0,
		expKeysSize: 0,
		keyStore:    newRdbKeyStore(),
		checksum:    [defaultRdbFileChecksumLength]byte{},
	}
}

func ReadRdbFile(iter *ByteIterator) (rdb *RdbFile, err error) {
	rdb = &RdbFile{}

	if rdb.header, err = readRdbHeader(iter); err != nil {
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

	if rdb.keyStore, err = readRdbFileValues(iter, rdb.hashSize); err != nil {
		return nil, fmt.Errorf("error reading rdb key values: %w", err)
	}

	if err = rdb.setChecksum(iter); err != nil {
		return nil, fmt.Errorf("error reading rdb checksum: %w", err)
	}

	return
}

func (f RdbFile) Iter() iter.Seq2[string, RdbKeyValue] {
	return maps.All(f.keyStore)
}

func (f *RdbFile) WriteContent(writer *bufio.Writer) (err error) {
	if _, headerErr := writer.Write(f.header.decode()); headerErr != nil {
		return fmt.Errorf("error writing header: %w", headerErr)
	}

	if eofErr := writer.WriteByte(EndOfFileEncoding); eofErr != nil {
		return fmt.Errorf("error writing enf of file: %w", eofErr)
	}

	if _, checksumErr := writer.Write(f.checksum[:]); checksumErr != nil {
		return fmt.Errorf("error writing checksum: %w", checksumErr)
	}

	return
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

	if markByte != SizesSectionEncoding {
		return fmt.Errorf(
			"expected %X mark, got %08b %X",
			SizesSectionEncoding,
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
	checkBytes, err := iter.readBytes(defaultRdbFileChecksumLength)
	if err != nil {
		return err
	}

	f.checksum = [8]byte(checkBytes)

	return
}
