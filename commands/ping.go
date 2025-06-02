package commands

import "github.com/codecrafters-io/redis-starter-go/rheltypes"

func CmdPing() (rheltypes.RhelType, error) {
	return rheltypes.SimpleString("PONG"), nil
}
