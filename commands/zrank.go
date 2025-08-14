package commands

import (
	"fmt"

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
	key := args.At(posZRankNameArg).String()
	item, found := GetDataMapInstance().Get(name)

	if !found {
		return rheltypes.NewNullBulkString(), nil
	}

	var set rheltypes.SortedSet

	var ok bool

	if set, ok = item.(rheltypes.SortedSet); !ok {
		return nil, c.ErrWrap(fmt.Errorf("expected sorted set, got %T", value))
	}

	return rheltypes.Integer(set.Index(key)), nil
}
