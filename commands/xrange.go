package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdXRange struct {
	BaseCommand
}

func NewCmdXRange() CmdXRange {
	return CmdXRange{BaseCommand: BaseCommand("XRANGE")}
}

const (
	posXRangeKey   = 0
	posXRangeLower = 1
	posXRangeUpper = 2
)

func (c CmdXRange) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	key := args.At(posXRangeKey).String()
	got, found := GetDataMapInstance().Get(key)

	if !found {
		return make(rheltypes.Array, 0), nil
	}

	var stream rheltypes.Stream

	var ok bool

	if stream, ok = got.(rheltypes.Stream); !ok {
		return nil, c.ErrWrap(fmt.Errorf("expected stream, got %T", value))
	}

	value = stream.Range(
		args.At(posXRangeLower).String(),
		args.At(posXRangeUpper).String(),
		true,
	).ToArray()

	return
}
