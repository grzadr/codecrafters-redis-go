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

const (
	numXReadStreamSections = 2
	numXReadValueSections  = 2

	// posXRangeKey   = 0
	// posXRangeLower = 1
	// posXRangeUpper = 2.
)

type CmdXReadStream struct {
	key string
	id  string
}

type CmdXReadArgs struct {
	block   int
	newOnly bool
	streams []CmdXReadStream
}

func NewCmdXReadArgs(args rheltypes.Array) (parsed CmdXReadArgs) {
	parsed.newOnly = args.At(-1).String() == "$"
	lastIndex := len(args)

	parsed.block = -1

	if parsed.newOnly {
		lastIndex--
	}

	args = args[:lastIndex]

	streamsIndex := -1

	var readInteger *int

	for i, a := range args[:lastIndex] {
		switch strings.ToUpper(a.String()) {
		case "BLOCK":
			readInteger = &parsed.block

			continue
		case "STREAMS":
			streamsIndex = i + 1
		}

		if readInteger != nil {
			*readInteger, _ = a.Integer()
			readInteger = nil
		}

		if streamsIndex != -1 {
			break
		}
	}

	args = args[streamsIndex:]

	half := len(args) / numXReadStreamSections

	parsed.streams = make([]CmdXReadStream, half)

	for i := range half {
		parsed.streams[i] = CmdXReadStream{
			key: args[i].String(),
			id:  args[i+half].String(),
		}
	}

	return parsed
}

func (c CmdXRead) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	parsedArgs := NewCmdXReadArgs(args)

	valueArray := make(rheltypes.Array, len(parsedArgs.streams))

	for s, streamSpec := range parsedArgs.streams {
		streamArray := make(rheltypes.Array, numXReadValueSections)

		streamArray[0] = rheltypes.NewBulkString(streamSpec.key)

		got, found := GetDataMapInstance().Get(streamSpec.key)

		if !found {
			return nil, c.ErrWrap(
				fmt.Errorf("stream %q not found", streamSpec.key),
			)
		}

		stream, ok := got.(rheltypes.Stream)

		if !ok {
			return nil, c.ErrWrap(fmt.Errorf("expected stream, got %T", value))
		}

		streamArray[1] = stream.Range(
			streamSpec.id,
			"+",
			false,
		).ToArray()

		valueArray[s] = streamArray
	}

	value = valueArray

	return value, err
}
