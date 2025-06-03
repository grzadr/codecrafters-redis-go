package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdEcho struct{}

func (CmdEcho) isRhelCommand() {}

func (CmdEcho) Name() string {
	return "ECHO"
}

func (c CmdEcho) ErrWrap(input error) (err error) {
	if input != nil {
		err = fmt.Errorf("failed to run %q command: %w", c.Name(), input)
	}
	return
}

func (e CmdEcho) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	value = args.First()
	if value == nil {
		err = e.ErrWrap(fmt.Errorf("expected message"))
	}
	return
}
