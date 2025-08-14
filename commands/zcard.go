package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdZCard struct {
	BaseCommand
}

func NewCmdZCard() CmdZCard {
	return CmdZCard{BaseCommand: BaseCommand("ZCARD")}
}

const (
	posZCardNameArg = 0
)

func (c CmdZCard) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	key := args.At(posZCardNameArg).String()
	item, found := GetDataMapInstance().Get(key)

	if !found {
		return rheltypes.Integer(0), nil
	}

	var set rheltypes.SortedSet

	var ok bool

	if set, ok = item.(rheltypes.SortedSet); !ok {
		return nil, c.ErrWrap(fmt.Errorf("expected stream, got %T", value))
	}

	return rheltypes.Integer(set.Size()), nil
}
