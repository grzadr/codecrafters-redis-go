package commands

import (
	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdReplconf struct {
	BaseCommand
}

func NewCmdReplconf() CmdReplconf {
	return CmdReplconf{BaseCommand: BaseCommand("REPLCONF")}
}

func (c CmdReplconf) Exec(args rheltypes.Array) (rheltypes.RhelType, error) {
	return rheltypes.SimpleString("PONG"), nil
}

func (c CmdReplconf) Render(name, value string) (cmd rheltypes.Array) {
	return rheltypes.NewArrayFromStrings(
		[]string{string(c.BaseCommand), name, value},
	)
}
