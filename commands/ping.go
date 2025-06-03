package commands

import (
	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdPing struct{}

func (CmdPing) isRhelCommand() {}

func (CmdPing) Name() string {
	return "PING"
}

func (CmdPing) Exec(args rheltypes.Array) (rheltypes.RhelType, error) {
	return rheltypes.SimpleString("PONG"), nil
}
