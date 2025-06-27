package commands

import (
	"fmt"
	"strconv"

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

	numInt := 0

	if !found {
		numInt = 1
	} else if numInt, err = num.Integer(); err != nil {
		return rheltypes.NewGenericError(
			fmt.Errorf("value is not an integer or out of range"),
		), nil
	} else {
		numInt++
	}

	value = rheltypes.Integer(numInt)

	instance.Set(key, rheltypes.NewBulkString(strconv.Itoa(numInt)))

	return value, err
}
