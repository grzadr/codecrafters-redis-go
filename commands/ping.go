package commands

import (
	"log"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdPing struct{}

func (CmdPing) isRhelCommand() {}

func (CmdPing) Name() string {
	return "PING"
}

func (CmdPing) Exec(value rheltypes.RhelType) (rheltypes.RhelType, error) {
	log.Println("Ping")
	return rheltypes.SimpleString("PONG"), nil
}
