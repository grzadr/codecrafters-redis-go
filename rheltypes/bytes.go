package rheltypes

// import "fmt"

// type Bytes struct {
// 	data []byte
// 	size int
// }

// func NewBytesFromTokens(i *TokenIterator) (b Bytes, err error) {
// 	b.size, err = i.LastToken.AsSize()
// 	if err != nil {
// 		return b, fmt.Errorf(
// 			"failed to extract size from %v: %w",
// 			i.LastToken,
// 			err,
// 		)
// 	}

// 	if b.size < 1 {
// 		return
// 	}

// 	d, err := i.readBytes(b.size)
// 	if err != nil {
// 		return
// 	}

// 	b.data = d

// 	return
// }

// func (b Bytes) Serialize() (buf []byte) {
// 	header := fmt.Sprintf("$%d\r\n", b.size)
// 	buf = make([]byte, 0, b.size+len(header))
// 	buf = fmt.Append(buf, header, b)

// 	return
// }

// func (b Bytes) String() string { return string(b) }

// func (b Bytes) Integer() (int, error) {
// 	return 0, nil
// }

// func (b Bytes) Size() int {
// 	return b.size
// }

// func (b Bytes) First() RhelType {
// 	return nil
// }

// func (b Bytes) isRhelType() {}
