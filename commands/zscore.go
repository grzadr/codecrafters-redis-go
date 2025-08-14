package commands

import (
	"fmt"
	"log"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdZScore struct {
	BaseCommand
}

func NewCmdZScore() CmdZScore {
	return CmdZScore{BaseCommand: BaseCommand("ZSCORE")}
}

const (
	posZScoreNameArg = 0
	posZSCoreKeyArg  = 1
)

func (c CmdZScore) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	name := args.At(posZScoreNameArg).String()
	key := args.At(posZSCoreKeyArg).String()
	item, found := GetDataMapInstance().Get(name)

	if !found {
		log.Println("item nit found", name)

		return rheltypes.NewNullBulkString(), nil
	}

	var set rheltypes.SortedSet

	var ok bool

	if set, ok = item.(rheltypes.SortedSet); !ok {
		return nil, c.ErrWrap(fmt.Errorf("expected sorted set, got %T", value))
	}

	member, found := set.Get(key)

	if !found {
		log.Println("key nit found", key)

		return rheltypes.NewNullBulkString(), nil
	}

	return member.AsBulkString(), nil
}
