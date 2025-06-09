package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdInfo struct {
	BaseCommand
}

func NewCmdInfo() CmdInfo {
	return CmdInfo{BaseCommand: BaseCommand("INFO")}
}

func (c CmdInfo) SubCmdReplication() (value rheltypes.RhelType, err error) {
	config := GetConfigMapInstance()

	_, ok := config.Get("replicaof")

	if !ok {
		return rheltypes.NewBulkString("role:master"), nil
	} else {
		return rheltypes.NewBulkString("role:slave"), nil
	}
}

func (c CmdInfo) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	subCmd := args.First()
	if subCmd == nil {
		return nil, c.ErrWrap(fmt.Errorf("expected sub command"))
	}

	switch subCmd.String() {
	case "replication":
		return c.SubCmdReplication()
	default:
		return nil, fmt.Errorf("unrecognized sub command %q", subCmd)
	}
}
