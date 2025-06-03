package commands

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type RhelCommand interface {
	isRhelCommand()
	String() string
}

type rhelCommand struct {
	Cmd string
}

func (c rhelCommand) isRhelCommand() {}

func (c rhelCommand) String() string {
	return c.Cmd
}

var (
	CmdUnknown = rhelCommand{}
	CmdPing    = rhelCommand{Cmd: "PING"}
	CmdEcho    = rhelCommand{Cmd: "ECHO"}
)

func cleanCommand(cmd rheltypes.RhelType) RhelCommand {
	switch cmd.String() {
	case "PING":
		return CmdPing
	case "ECHO":
		return CmdEcho
	default:
		return CmdUnknown
	}
}

func ExecuteCommand(content []byte) (rheltypes.RhelType, error) {
	log.Printf("Command %q:\n%s", content, hex.Dump(content))
	if len(content) == 0 {
		return nil, fmt.Errorf("missing command")
	}

	items, err := rheltypes.NewArray(content)
	if err != nil {
		return nil, err
	}

	cmd := items[0]
	rest := items[1:]

	switch cleanCommand(cmd) {
	case CmdPing:
		return RunPing()
	case CmdEcho:
		return RunEcho(rest)
	default:
		return nil, fmt.Errorf("unknown command %s", cmd)
	}
}
