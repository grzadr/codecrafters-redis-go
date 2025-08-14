package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdZRange struct {
	BaseCommand
}

func NewCmdZRange() CmdZRange {
	return CmdZRange{BaseCommand: BaseCommand("ZRANGE")}
}

const (
	posZRangeNameArg  = 0
	posZRangeStartArg = 1
	posZRangeStopArg  = 2
)

func (c CmdZRange) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	key := args.At(posZRangeNameArg).String()
	item, found := GetDataMapInstance().Get(key)

	if !found {
		return make(rheltypes.Array, 0), nil
	}

	var stream rheltypes.SortedSet

	var ok bool

	if stream, ok = item.(rheltypes.SortedSet); !ok {
		return nil, c.ErrWrap(fmt.Errorf("expected stream, got %T", value))
	}

	start, _ := args.At(posZRangeStartArg).Integer()
	stop, _ := args.At(posZRangeStopArg).Integer()

	value = stream.Range(
		start,
		stop,
	)

	return
}
