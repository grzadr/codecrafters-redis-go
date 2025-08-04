package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/pubsub"
	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdRPushArgs struct {
	Key   string
	Items []rheltypes.RhelType
}

func NewCmdRLPushArgs(args rheltypes.Array) (parsed CmdRPushArgs, err error) {
	parsed.Key = args.At(0).String()
	parsed.Items = make([]rheltypes.RhelType, len(args)-1)

	if len(args) > 1 {
		copy(parsed.Items, args[1:])
	}

	return
}

type CmdRPush struct {
	BaseCommand
}

func NewCmdRPush() CmdRPush {
	return CmdRPush{BaseCommand: BaseCommand("RPUSH")}
}

func (c CmdRPush) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	parsedArgs, err := NewCmdRLPushArgs(args)
	if err != nil {
		return nil, c.ErrWrap(err)
	}

	instance := GetDataMapInstance()

	value, found := instance.Get(parsedArgs.Key)

	var list rheltypes.Array

	var ok bool

	if !found {
		list = make(rheltypes.Array, 0)
	} else if list, ok = value.(rheltypes.Array); !ok {
		return nil, c.ErrWrap(fmt.Errorf("expected stream, got %T", value))
	}

	sm := pubsub.GetStreamManager()

	for _, item := range parsedArgs.Items {
		list = append(list, item)

		go sm.Publish(parsedArgs.Key, item)
	}

	instance.Set(parsedArgs.Key, list)

	return rheltypes.Integer(len(list)), nil
}
