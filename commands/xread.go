package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/pubsub"
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
	streams []CmdXReadStream
}

func NewCmdXReadArgs(args rheltypes.Array) (parsed CmdXReadArgs) {
	parsed.block = -1

	streamsIndex := -1

	var readInteger *int

	for i, a := range args {
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

func (c CmdXRead) ReadAll(
	streams []CmdXReadStream,
) (values rheltypes.Array, err error) {
	values = make(rheltypes.Array, len(streams))

	for s, streamSpec := range streams {
		if streamSpec.id == "$" {
			continue
		}

		streamArray := make(rheltypes.Array, numXReadValueSections)

		streamArray[0] = rheltypes.NewBulkString(streamSpec.key)

		got, found := GetDataMapInstance().Get(streamSpec.key)

		if !found {
			return nil, fmt.Errorf("stream %q not found", streamSpec.key)
		}

		stream, ok := got.(rheltypes.Stream)

		if !ok {
			return nil, fmt.Errorf(
				"expected stream from %q, got %T %v",
				streamSpec.key,
				got,
				got,
			)
		}

		streamArray[1] = stream.Range(
			streamSpec.id,
			"+",
			false,
		).ToArray()

		values[s] = streamArray
	}

	return values, err
}

func createContextFromTimeout(
	timeoutMS int,
) (context.Context, context.CancelFunc) {
	switch {
	case timeoutMS == 0:
		// Infinite timeout with cancellation capability
		return context.WithCancel(context.Background())
	case timeoutMS > 0:
		// Explicit timeout duration
		duration := time.Duration(timeoutMS) * time.Millisecond

		return context.WithTimeout(context.Background(), duration)
	default:
		// Invalid negative values default to no timeout
		return context.Background(), func() {}
	}
}

func (c CmdXRead) ReadLast(
	key string, timeout int,
) (value rheltypes.StreamItem, err error) {
	sub := pubsub.GetStreamManager().Subscribe(key)
	defer sub.Close()

	ticker := time.NewTicker(defaultWaitTicker)
	defer ticker.Stop()

	ctx, cancel := createContextFromTimeout(timeout)
	defer cancel()

	for {
		select {
		case msg := <-sub.Messages:
			stream, ok := msg.(*rheltypes.StreamItem)

			if !ok {
				return value, fmt.Errorf(
					"expected stream, got %T %v",
					msg,
					msg,
				)
			}

			return *stream, nil

		case <-ctx.Done():
			return value, nil
		}
	}
}

func (c CmdXRead) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	parsedArgs := NewCmdXReadArgs(args)

	valueArray := make(rheltypes.Array, 0, len(parsedArgs.streams))

	if parsedArgs.block == -1 {
		if valueArray, err = c.ReadAll(parsedArgs.streams); err != nil {
			return nil, c.ErrWrap(fmt.Errorf("failed to read all: %w", err))
		}
	}

	if parsedArgs.block > -1 {
		key := parsedArgs.streams[0].key

		last, err := c.ReadLast(key, parsedArgs.block)
		if err != nil {
			return nil, c.ErrWrap(fmt.Errorf("failed to read last: %w", err))
		} else if last.Size() == 0 {
			return rheltypes.NewNullBulkString(), nil
		}

		lastArray := last.ToArray()

		if len(valueArray) > 0 {
			streamArray := valueArray.At(0).(rheltypes.Array)
			stream := streamArray.At(1).(rheltypes.Array)
			stream.Append(lastArray)
			streamArray.Set(1, stream)
			valueArray.Set(0, streamArray)
		} else {
			valueArray = rheltypes.Array{
				rheltypes.Array{
					rheltypes.NewBulkString(key),
					rheltypes.Array{lastArray},
				},
			}
		}
	}

	if len(valueArray) == 0 {
		return rheltypes.NewNullBulkString(), nil
	}

	value = valueArray

	return value, err
}
