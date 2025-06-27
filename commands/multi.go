package commands

import (
	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdMulti struct {
	BaseCommand
}

func NewCmdMulti() CmdMulti {
	return CmdMulti{BaseCommand: BaseCommand("MULTI")}
}

func (c CmdMulti) Exec(args rheltypes.Array) (rheltypes.RhelType, error) {
	return rheltypes.SimpleString("OK"), nil
}
