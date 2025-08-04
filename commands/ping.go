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

func (c CmdPing) Exec(args rheltypes.Array) (rheltypes.RhelType, error) {
	return rheltypes.SimpleString("PONG"), nil
}

func (c CmdPing) Render() (cmd rheltypes.Array) {
	return rheltypes.NewArrayFromStrings([]string{string(c.BaseCommand)})
}

func (c CmdPing) AllowedInSubscription() bool { return true }
