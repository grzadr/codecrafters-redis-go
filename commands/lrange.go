package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdLRange struct {
	BaseCommand
}

func NewCmdLRange() CmdLRange {
	return CmdLRange{BaseCommand: BaseCommand("LRANGE")}
}

func (c CmdLRange) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	key := args.At(0).String()
	start, _ := args.At(1).Integer()
	stop, _ := args.At(2).Integer()

	if start > stop {
		return rheltypes.Array{}, nil
	}

	instance := GetDataMapInstance()

	value, found := instance.Get(key)

	var list rheltypes.Array

	var ok bool

	if !found {
		return rheltypes.Array{}, nil
	} else if list, ok = value.(rheltypes.Array); !ok {
		return nil, c.ErrWrap(fmt.Errorf("expected stream, got %T", value))
	}

	return list.Range(start, stop), nil
}
