package commands

import (
	"fmt"
	"log"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdZRank struct {
	BaseCommand
}

func NewCmdZRank() CmdZRank {
	return CmdZRank{BaseCommand: BaseCommand("ZRANK")}
}

const (
	posZRankNameArg = 0
	posZRankKeyArg  = 1
)

func (c CmdZRank) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	name := args.At(posZRankNameArg).String()
	key := args.At(posZRankKeyArg).String()
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

	index, found := set.Index(key)

	if !found {
		log.Println("key nit found", key)

		return rheltypes.NewNullBulkString(), nil
	}

	return rheltypes.Integer(index), nil
}
