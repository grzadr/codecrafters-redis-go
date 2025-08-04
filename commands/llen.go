package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdLLen struct {
	BaseCommand
}

func NewCmdLLen() CmdLLen {
	return CmdLLen{BaseCommand: BaseCommand("LLEN")}
}

func (c CmdLLen) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	key := args.At(0).String()

	instance := GetDataMapInstance()

	if value, found := instance.Get(key); !found {
		return rheltypes.Integer(0), nil
	} else if list, ok := value.(rheltypes.Array); !ok {
		return nil, c.ErrWrap(fmt.Errorf("expected stream, got %T", value))
	} else {
		return rheltypes.Integer(len(list)), nil
	}
}
