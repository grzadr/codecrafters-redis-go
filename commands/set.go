package commands

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdSetArgs struct {
	Key   string
	Value rheltypes.RhelType
	Px    int
}

func (a *CmdSetArgs) Set(name string, value rheltypes.RhelType) (err error) {
	switch name {
	case "PX":
		if a.Px, err = value.Integer(); err != nil {
			return fmt.Errorf(
				"failed to convert value %q to integer: %w",
				value.String(),
				err,
			)
		}

	default:
		return fmt.Errorf("unknown set option %q", name)
	}

	return nil
}

type CmdSet struct {
	BaseCommand
}

func NewCmdSet() CmdSet {
	return CmdSet{BaseCommand: BaseCommand("SET")}
}

func parseSetArgs(args rheltypes.Array) (parsed CmdSetArgs, err error) {
	setKey := args.At(0)
	if setKey == nil {
		return parsed, fmt.Errorf("missing key")
	}

	parsed.Key = setKey.String()
	setValue := args.At(1)

	if setValue == nil {
		return parsed, fmt.Errorf("missing value")
	}

	parsed.Value = setValue

	lastField := ""

	for _, arg := range args[2:] {
		switch name := strings.ToUpper(arg.String()); name {
		case "PX":
			lastField = name
		default:
			if lastField == "" {
				continue
			}

			err = parsed.Set(lastField, arg)
			if err != nil {
				return parsed, err
			}

			lastField = ""
		}
	}

	return parsed, err
}

func (c CmdSet) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	parsedArgs, err := parseSetArgs(args)
	if err != nil {
		return nil, c.ErrWrap(err)
	}

	instance := GetDataMapInstance()

	if parsedArgs.Px > 0 {
		instance.SetToExpire(
			parsedArgs.Key,
			parsedArgs.Value,
			int64(parsedArgs.Px),
		)
	} else {
		instance.Set(parsedArgs.Key, parsedArgs.Value)
	}

	return rheltypes.SimpleString("OK"), nil
}

func (c CmdSet) Resend() bool { return true }

func (c CmdSet) ReplicaRespond() bool { return true }
