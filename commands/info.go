package commands

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdInfo struct {
	BaseCommand
}

func NewCmdInfo() CmdInfo {
	return CmdInfo{BaseCommand: BaseCommand("INFO")}
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
		return c.subCmdReplication()
	default:
		return nil, fmt.Errorf("unrecognized sub command %q", subCmd)
	}
}

func (c CmdInfo) subCmdReplication() (value rheltypes.RhelType, err error) {
	config := GetConfigMapInstance()

	fields := []string{"role", "master_replid", "master_repl_offset"}
	str := make([]string, 0, len(fields))

	for _, key := range fields {
		value, ok := config.Get(key)
		if !ok {
			continue
		}

		str = append(str, fmt.Sprintf("%s:%s", key, value))
	}

	value = rheltypes.NewBulkString(strings.Join(str, "\n"))

	return
}
