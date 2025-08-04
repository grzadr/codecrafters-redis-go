package commands

import (
	"github.com/codecrafters-io/redis-starter-go/pubsub"
	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

const cmdSubscribeResultNumPos = 2

type CmdSubscribe struct {
	BaseCommand
}

func NewCmdSubscribe() CmdSubscribe {
	return CmdSubscribe{BaseCommand: BaseCommand("SUBSCRIBE")}
}

func (c CmdSubscribe) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	key := args.At(0).String()

	sub := pubsub.GetStreamManager().Subscribe(key, false)

	// // timeout, _ := args.At(1).Float()
	// last, err := pubsub.ReadLast(key, int(timeout*milisecondInSecond))
	// if err != nil {
	// 	return nil, c.ErrWrap(fmt.Errorf("failed to read last: %w", err))
	// } else if last == nil {
	// 	log.Println("returning null string")
	// 	return rheltypes.NewNullBulkString(), nil
	// } else {
	// 	return rheltypes.Array{
	// 		rheltypes.NewBulkString(key),
	// 		last,
	// 	}, nil
	// }
	arr := rheltypes.NewArrayFromStrings([]string{"subscribe", key})
	value = append(arr, rheltypes.Integer(sub.Id))

	return value, nil
}

func (c CmdSubscribe) AllowedInSubscription() bool { return true }
