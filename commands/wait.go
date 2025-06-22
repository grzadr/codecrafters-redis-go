package commands

import (
	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdWait struct {
	BaseCommand
}

func NewCmdWait() CmdWait {
	return CmdWait{BaseCommand: BaseCommand("WAIT")}
}

func (c CmdWait) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	return rheltypes.Integer(0), nil
}

func (c CmdWait) Render(id, offset string) (cmd rheltypes.Array) {
	return rheltypes.NewArrayFromStrings(
		[]string{string(c.BaseCommand), id, offset},
	)
}
