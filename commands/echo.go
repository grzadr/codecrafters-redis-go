package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdEcho struct {
	BaseCommand
}

func NewCmdEcho() CmdEcho {
	return CmdEcho{BaseCommand: BaseCommand("ECHO")}
}

func (c CmdEcho) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	value = args.First()
	if value == nil {
		err = c.ErrWrap(fmt.Errorf("expected message"))
	}

	return
}
