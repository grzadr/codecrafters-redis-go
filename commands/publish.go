package commands

import (
	"github.com/codecrafters-io/redis-starter-go/pubsub"
	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdPublish struct {
	BaseCommand
}

func NewCmdPublish() CmdPublish {
	return CmdPublish{BaseCommand: BaseCommand("PUBLISH")}
}

func (c CmdPublish) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	key := args.At(0).String()
	msg := args.At(1)

	sm := pubsub.GetStreamManager()

	value = rheltypes.Integer(sm.NumSubscribers(key))

	go sm.Publish(key, msg)

	return value, nil
}

func (c CmdPublish) AllowedInSubscription() bool { return true }
