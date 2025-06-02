package commands

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

func ExecuteCommand(content []byte) (rheltypes.RhelType, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("missing command")
	}
	cmd, rest, _ := strings.Cut(content, " ")

	switch cmd := strings.ToUpper(cmd); cmd {
	case "PING":
		return CmdPing()
	case "ECHO":
		return CmdEcho(rest)
	default:
		return nil, fmt.Errorf("unknown command %s", cmd)
	}
}
