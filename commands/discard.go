package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdDiscard struct {
	BaseCommand
}

func NewCmdDiscard() CmdDiscard {
	return CmdDiscard{BaseCommand: BaseCommand("DISCARD")}
}

func (c CmdDiscard) Exec(args rheltypes.Array) (rheltypes.RhelType, error) {
	if args == nil {
		return rheltypes.NewGenericError(fmt.Errorf("EXEC without MULTI")), nil
	} else {
		return args, nil
	}
}
