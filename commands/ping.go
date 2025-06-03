package commands

import (
	"log"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

func RunPing() (rheltypes.RhelType, error) {
	log.Println("Ping")
	return rheltypes.SimpleString("PONG"), nil
}
