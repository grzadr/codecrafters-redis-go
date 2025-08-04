package commands

import (
	"fmt"
	"slices"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdLPush struct {
	BaseCommand
}

func NewCmdLPush() CmdLPush {
	return CmdLPush{BaseCommand: BaseCommand("LPUSH")}
}

func (c CmdLPush) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	parsedArgs, err := NewCmdRLPushArgs(args)
	if err != nil {
		return nil, c.ErrWrap(err)
	}

	instance := GetDataMapInstance()

	value, found := instance.Get(parsedArgs.Key)

	var list rheltypes.Array

	var ok bool

	if !found {
		list = make(rheltypes.Array, 0)
	} else if list, ok = value.(rheltypes.Array); !ok {
		return nil, c.ErrWrap(fmt.Errorf("expected stream, got %T", value))
	}

	updated := make(rheltypes.Array, len(list)+len(parsedArgs.Items))

	for _, item := range slices.Backward(parsedArgs.Items) {
		updated = append(updated, item)
	}

	copy(updated[len(parsedArgs.Items):], list)

	instance.Set(parsedArgs.Key, updated)

	return rheltypes.Integer(len(updated)), nil
}
