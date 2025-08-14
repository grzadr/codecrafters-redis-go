package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

const (
	cmdZRemNameArg = 0
	cmdZRemKeyArg  = 1
)

type CmdZRem struct {
	BaseCommand
}

func NewCmdZRem() CmdZRem {
	return CmdZRem{BaseCommand: BaseCommand("ZREM")}
}

func (c CmdZRem) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	name := args.At(cmdZRemNameArg).String()
	key := args.At(cmdZRemKeyArg).String()
	instance := GetDataMapInstance()

	item, found := GetDataMapInstance().Get(name)

	if !found {
		return make(rheltypes.Array, 0), nil
	}

	var set rheltypes.SortedSet

	var ok bool

	if set, ok = item.(rheltypes.SortedSet); !ok {
		return nil, c.ErrWrap(fmt.Errorf("expected sorted set, got %T", value))
	}

	if set.Delete(key) {
		instance.Set(name, set)

		return rheltypes.Integer(1), nil
	}

	return rheltypes.Integer(0), nil
}
