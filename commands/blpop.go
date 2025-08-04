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

// func (c CmdBLPop) ReadLast(
// 	key string, timeout int,
// ) (rheltypes.RhelType, error) {
// 	sub := pubsub.GetStreamManager().Subscribe(key, true)
// 	defer sub.Close()

// 	// ticker := time.NewTicker(defaultWaitTicker)
// 	// defer ticker.Stop()

// 	ctx, cancel := pubreateContextFromTimeout(timeout)
// 	defer cancel()

// 	for {
// 		select {
// 		case msg := <-sub.Messages:
// 			item, ok := msg.(rheltypes.RhelType)

// 			if !ok {
// 				return nil, fmt.Errorf(
// 					"expected rheltype, got %T %v",
// 					msg,
// 					msg,
// 				)
// 			}

// 			return item, nil

// 		case <-ctx.Done():
// 			return nil, nil
// 		}
// 	}
// }

func (c CmdBLPop) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	key := args.At(0).String()

	timeout, _ := args.At(1).Float()

	lastMsg, err := pubsub.ReadLast(key, int(timeout*milisecondInSecond))

	last := lastMsg.(rheltypes.RhelType)

	if err != nil {
		return nil, c.ErrWrap(fmt.Errorf("failed to read last: %w", err))
	} else if last == nil {
		log.Println("returning null string")

		return rheltypes.NewNullBulkString(), nil
	} else {
		return rheltypes.Array{
			rheltypes.NewBulkString(key),
			last,
		}, nil
	}
}
