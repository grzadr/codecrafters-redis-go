package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdKeys struct {
	BaseCommand
}

func NewCmdKeys() CmdKeys {
	return CmdKeys{BaseCommand: BaseCommand("KEYS")}
}

func (c CmdKeys) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	key := args.At(0)
	if key == nil {
		return nil, c.ErrWrap(fmt.Errorf("missing key"))
	}

	instance := GetDataMapInstance()

	switch query := key.String(); query {
	case "*":
		buf := make(rheltypes.Array, instance.Size())
		i := 0

		for k := range instance.Keys() {
			buf[i] = rheltypes.NewBulkString(k)
			i++
		}

		return buf, nil
	default:
		return nil, c.ErrWrap(fmt.Errorf("unsupported syntax: %q", query))
	}
}
