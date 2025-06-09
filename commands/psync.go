package commands

import (
	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdPsync struct {
	BaseCommand
}

func NewCmdPsync() CmdPsync {
	return CmdPsync{BaseCommand: BaseCommand("PSYNC")}
}

func (c CmdPsync) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	return
}

func (c CmdPsync) Render(id, offset string) (cmd rheltypes.Array) {
	return rheltypes.NewArrayFromStrings(
		[]string{string(c.BaseCommand), id, offset},
	)
}
