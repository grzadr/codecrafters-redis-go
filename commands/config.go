package commands

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdConfig struct {
	BaseCommand
}

func NewCmdConfig() CmdConfig {
	return CmdConfig{BaseCommand: BaseCommand("CONFIG")}
}

const defaultGetValueLength = 2

func (c CmdConfig) Get(
	args rheltypes.Array,
) (value rheltypes.Array, err error) {
	key := args.At(0)
	if key == nil {
		return nil, c.ErrWrap(fmt.Errorf("missing get key"))
	}

	config := GetConfigMapInstance()

	value = append(make(rheltypes.Array, 0, defaultGetValueLength), key)

	if foundValue, found := config.Get(key.String()); found {
		value = append(value, foundValue)
	}

	return
}

func (c CmdConfig) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	cmd := args.At(0)
	if cmd == nil {
		return nil, c.ErrWrap(fmt.Errorf("missing config command"))
	}

	switch subcmd := strings.ToUpper(cmd.String()); subcmd {
	case "GET":
		return c.Get(args[1:])
	default:
		return nil, c.ErrWrap(fmt.Errorf("unknown config command %s", subcmd))
	}
}
