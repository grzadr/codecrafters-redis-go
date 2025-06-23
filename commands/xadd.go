package commands

import (
	"fmt"
	"slices"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

const defaultXAddSliceSize = 2

type CmdXAddArgs struct {
	Key   string
	Id    string
	Items map[string]string
}

func NewXAddArgs(args rheltypes.Array) (parsed CmdXAddArgs, err error) {
	parsed.Key = args.At(0).String()
	parsed.Id = args.At(1).String()
	items := args[2:]
	parsed.Items = make(map[string]string, len(items)/defaultXAddSliceSize)

	for pair := range slices.Chunk(items, defaultXAddSliceSize) {
		parsed.Items[pair[0].String()] = pair[1].String()
	}

	return
}

type CmdXAdd struct {
	BaseCommand
}

func NewCmdXAdd() CmdXAdd {
	return CmdXAdd{BaseCommand: BaseCommand("XADD")}
}

func (c CmdXAdd) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	parsedArgs, err := NewXAddArgs(args)
	if err != nil {
		return nil, c.ErrWrap(err)
	}

	instance := GetDataMapInstance()

	value, found := instance.Get(parsedArgs.Key)

	var stream rheltypes.Stream

	var ok bool

	if !found {
		stream = rheltypes.NewStream()
	} else if stream, ok = value.(rheltypes.Stream); !ok {
		return nil, c.ErrWrap(fmt.Errorf("expected stream, got %T", value))
	}

	var addedId string

	if addedId, err = stream.Add(parsedArgs.Id, parsedArgs.Items); err != nil {
		return rheltypes.NewGenericError(err), nil
	}

	instance.Set(parsedArgs.Key, stream)

	return rheltypes.NewBulkString(addedId), nil
}
