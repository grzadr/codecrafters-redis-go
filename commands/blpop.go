package commands

import (
	"fmt"
	"log"

	"github.com/codecrafters-io/redis-starter-go/pubsub"
	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdBLPop struct {
	BaseCommand
}

func NewCmdBLPop() CmdBLPop {
	return CmdBLPop{BaseCommand: BaseCommand("BLPOP")}
}

const milisecondInSecond = 1000

func (c CmdBLPop) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	key := args.At(0).String()

	timeout, _ := args.At(1).Float()

	lastMsg, err := pubsub.ReadLast(key, int(timeout*milisecondInSecond))

	if err != nil {
		return nil, c.ErrWrap(fmt.Errorf("failed to read last: %w", err))
	} else if lastMsg == nil {
		log.Println("returning null string")

		return rheltypes.NewNullBulkString(), nil
	}

	last := lastMsg.(rheltypes.RhelType)

	return rheltypes.Array{
		rheltypes.NewBulkString(key),
		last,
	}, nil
}
