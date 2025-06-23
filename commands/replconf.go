package commands

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/connection"
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
	switch args.Cmd() {
	case "GETACK":
		offset := connection.GetOffsetTracker().Current()

		return c.Render("ACK", strconv.Itoa(offset)), nil
	case "ACK":
		return nil, nil
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
