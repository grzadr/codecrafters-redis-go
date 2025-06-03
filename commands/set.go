package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdSet struct{}

func (CmdSet) isRhelCommand() {}

func (CmdSet) Name() string {
	return "SET"
}

func (c CmdSet) ErrWrap(input error) (err error) {
	if input != nil {
		err = fmt.Errorf("failed to run %q command: %w", c.Name(), input)
	}
	return
}

func (c CmdSet) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	setKey := args.At(0)
	if setKey == nil {
		return nil, c.ErrWrap(fmt.Errorf("missing key"))
	}
	setValue := args.At(1)
	if setValue == nil {
		return nil, c.ErrWrap(fmt.Errorf("missing value"))
	}
	instance := rheltypes.GetSageMapInstance()

	instance.Set(setKey.String(), setValue)
	return rheltypes.SimpleString("OK"), nil
}
