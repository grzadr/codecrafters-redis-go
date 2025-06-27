package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdIncr struct {
	BaseCommand
}

func NewCmdIncr() CmdIncr {
	return CmdIncr{BaseCommand: BaseCommand("INCR")}
}

func (c CmdIncr) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	key := args.First().String()

	instance := GetDataMapInstance()

	num, found := instance.Get(key)

	if !found {
		value = rheltypes.Integer(1)
	} else if numInt, ok := num.(rheltypes.Integer); !ok {
		return rheltypes.NewGenericError(fmt.Errorf("value is not an integer or out of range")), nil
	} else {
		value = rheltypes.Integer(numInt + 1)
	}

	instance.Set(key, value)

	return value, err
}
