package rheltypes

import (
	"fmt"
)

type ErrorType string

const (
	GenericErrorType = "ERR"
)

type Error struct {
	errType ErrorType
	msg     string
}

func NewGenericError(msg error) Error {
	return Error{errType: GenericErrorType, msg: msg.Error()}
}

// func NewSimpleStringFromTokens(token Token) (SimpleString, error) {
// 	return SimpleString(token.Data), nil
// }

func (e Error) Size() int {
	return len(
		ErrorPrefix,
	) + len(
		e.errType,
	) + len(
		" ",
	) + len(
		e.msg,
	) + len(
		rhelFieldDelim,
	)
}

func (e Error) Serialize() []byte {
	buf := make([]byte, 0, e.Size())

	return fmt.Appendf(
		buf,
		"%s%s %s%s",
		ErrorPrefix,
		e.errType,
		e.msg,
		rhelFieldDelim,
	)
}

func (e Error) String() string {
	return string(e.Serialize())
}

func (e Error) First() RhelType {
	return e
}

func (e Error) Integer() (num int, err error) {
	return 0, nil
}

func (e Error) TypeName() string {
	return "error"
}

func (e Error) Float() (float64, error) { return 0, nil }

func (e Error) isRhelType() {}
