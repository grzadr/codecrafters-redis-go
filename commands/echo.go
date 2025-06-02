package commands

import "github.com/codecrafters-io/redis-starter-go/rheltypes"

func CmdEcho(content string) (rheltypes.RhelType, error) {
	return rheltypes.BulkString(content), nil
}
