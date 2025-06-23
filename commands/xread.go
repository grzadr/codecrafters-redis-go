package commands

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdXRead struct {
	BaseCommand
}

func NewCmdXRead() CmdXRead {
	return CmdXRead{BaseCommand: BaseCommand("XREAD")}
}

// const (
// 	posXRangeKey   = 0
// 	posXRangeLower = 1
// 	posXRangeUpper = 2
// )

type CmdXReadStreams struct {
	key string
	id string
}

type CmdXReadArgs struct {
	block int
	newOnly bool
	streams []CmdXReadStreams
}

func NewCmdXReadArgs(args rheltypes.Array) (parsed CmdXReadArgs) {
	parsed.newOnly = args.At(-1).String() == "$"
	lastIndex := len(args)

	parsed.block = -1

	if parsed.newOnly {
		lastIndex--
	}

	readBlock := false

	args = args[:lastIndex]

	streamsIndex := 0

	for i, a := range args[:lastIndex] {
		switch strings.ToUpper(a) {
		case "BLOCK":
			readBlock = true
			continue
		case "STREAMS":
			streamsIndex = i + 1
			break
		}
		if readBlock {
			parsed.block = a.Integer()
		}
	}



}

func (c CmdXRead) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {

	parsedArgs =

	key := args.At(posXRangeKey).String()
	got, found := GetDataMapInstance().Get(key)

	if !found {
		return make(rheltypes.Array, 0), nil
	}

	var stream rheltypes.Stream

	var ok bool

	if stream, ok = got.(rheltypes.Stream); !ok {
		return nil, c.ErrWrap(fmt.Errorf("expected stream, got %T", value))
	}

	value = stream.Range(
		args.At(posXRangeLower).String(),
		,
	).ToArray()

	return
}
