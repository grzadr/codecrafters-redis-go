package commands

import (
	"log"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

func RunEcho(content []rheltypes.RhelType) (rheltypes.RhelType, error) {
	log.Printf("Echo %q", content)
	return rheltypes.Array(content), nil
}
