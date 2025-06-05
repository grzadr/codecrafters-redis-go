package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdGet struct{}

func (CmdGet) Name() string {
	return "SET"
}

func (c CmdGet) ErrWrap(input error) (err error) {
	if input != nil {
		err = fmt.Errorf("failed to run %q command: %w", c.Name(), input)
	}

	return
}

func (c CmdGet) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	key := args.At(0)
	if key == nil {
		return nil, c.ErrWrap(fmt.Errorf("missing key"))
	}

	instance := GetDataMapInstance()

	var found bool

	value, found = instance.Get(key.String())
	if !found {
		value = rheltypes.NewNullBulkString()
	}

	return
}

func (CmdGet) isRhelCommand() {}
