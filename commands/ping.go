package commands

import (
	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdPing struct {
	BaseCommand
}

func NewCmdPing() CmdPing {
	return CmdPing{BaseCommand: BaseCommand("PING")}
}

func (CmdPing) Exec(args rheltypes.Array) (rheltypes.RhelType, error) {
	return rheltypes.SimpleString("PONG"), nil
}
