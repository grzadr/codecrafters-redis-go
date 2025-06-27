package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdExec struct {
	BaseCommand
}

func NewCmdExec() CmdExec {
	return CmdExec{BaseCommand: BaseCommand("EXEC")}
}

func (c CmdExec) Exec(args rheltypes.Array) (rheltypes.RhelType, error) {
	return rheltypes.NewGenericError(fmt.Errorf("EXEC without MULTI")), nil
}
