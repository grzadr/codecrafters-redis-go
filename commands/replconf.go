package commands

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdReplconf struct {
	BaseCommand
}

func NewCmdReplconf() CmdReplconf {
	return CmdReplconf{BaseCommand: BaseCommand("REPLCONF")}
}

func (c CmdReplconf) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	subcommand := strings.ToUpper(args.First().String())
	switch subcommand {
	case "GETACK":
		return c.Render("ACK", "0"), nil
	default:
		return rheltypes.NewBulkString("OK"), nil
	}
}

func (c CmdReplconf) Render(name, value string) (cmd rheltypes.Array) {
	return rheltypes.NewArrayFromStrings(
		[]string{string(c.BaseCommand), name, value},
	)
}

func (c CmdReplconf) ReplicaRespond() bool { return true }
