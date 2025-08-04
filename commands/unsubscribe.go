package commands

import (
	"github.com/codecrafters-io/redis-starter-go/pubsub"
	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdUnsubscribe struct {
	BaseCommand
}

func NewCmdUnsubscribe() CmdUnsubscribe {
	return CmdUnsubscribe{BaseCommand: BaseCommand("UNSUBSCRIBE")}
}

func (c CmdUnsubscribe) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	key := args.At(0).String()
	id, _ := args.At(1).Integer()

	pubsub.GetStreamManager().Unsubscribe(key, id)

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
	arr := rheltypes.NewArrayFromStrings([]string{"unsubscribe", key})

	return append(arr, rheltypes.Integer(0)), nil
}

func (c CmdUnsubscribe) AllowedInSubscription() bool { return true }
