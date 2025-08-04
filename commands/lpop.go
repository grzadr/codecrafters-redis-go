package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdLPop struct {
	BaseCommand
}

func NewCmdLPop() CmdLPop {
	return CmdLPop{BaseCommand: BaseCommand("LPOP")}
}

func (c CmdLPop) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	key := args.At(0).String()

	start := 0

	if startArg := args.At(1); startArg != nil {
		start, _ = startArg.Integer()
	}

	instance := GetDataMapInstance()

	value, found := instance.Get(key)

	var list rheltypes.Array

	var ok bool

	if !found {
		return rheltypes.NewBulkString("-1"), nil
	} else if list, ok = value.(rheltypes.Array); !ok {
		return nil, c.ErrWrap(fmt.Errorf("expected stream, got %T", value))
	}

	if len(list) == 0 {
		return rheltypes.NewBulkString("-1"), nil
	}

	var output rheltypes.RhelType

	if start == 0 {
		output = list.At(0)
		list = list[1:]
	} else {
		output = list.Range(0, start)
		list = list[min(start, len(list)):]
	}

	instance.Set(key, list)

	return output, nil
}
