package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdXAddArgs struct {
	Key   string
	Id    string
	Items rheltypes.Array
}

func NewXAddArgs(args rheltypes.Array) (parsed CmdXAddArgs, err error) {
	// 	setKey := args.At(0)
	// 	if setKey == nil {
	// 		return parsed, fmt.Errorf("missing key")
	// 	}
	parsed.Key = args.At(0).String()
	parsed.Id = args.At(1).String()
	parsed.Items = args[2:]

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

	item := rheltypes.NewStreamItemFromArray(parsedArgs.Id, parsedArgs.Items)
	instance := GetDataMapInstance()

	value, found := instance.Get(parsedArgs.Key)

	var stream rheltypes.Stream

	var ok bool

	if !found {
		stream = rheltypes.NewStream()
	} else if stream, ok = value.(rheltypes.Stream); !ok {
		return nil, c.ErrWrap(fmt.Errorf("expected stream, got %T", value))
	}

	stream = append(stream, item)

	instance.Set(parsedArgs.Key, stream)

	return rheltypes.NewBulkString(parsedArgs.Id), nil
}
