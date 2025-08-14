package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

// type CmdRPushArgs struct {
// 	Key   string
// 	Items []rheltypes.RhelType
// }

// func NewCmdRLPushArgs(args rheltypes.Array) (parsed CmdRPushArgs, err error)
// {
// 	parsed.Key = args.At(0).String()
// 	parsed.Items = make([]rheltypes.RhelType, len(args)-1)

// 	if len(args) > 1 {
// 		copy(parsed.Items, args[1:])
// 	}

// 	return
// }

const (
	cmdZAddNameArg  = 0
	cmdZAddKeyArg   = 1
	cmdZAddScoreArg = 2
)

type CmdZAdd struct {
	BaseCommand
}

func NewCmdZAdd() CmdZAdd {
	return CmdZAdd{BaseCommand: BaseCommand("ZADD")}
}

func (c CmdZAdd) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	name := args.At(cmdZAddNameArg).String()
	score, _ := args.At(cmdZAddKeyArg).Float()
	key := args.At(cmdZAddScoreArg).String()
	instance := GetDataMapInstance()

	var set rheltypes.SortedSet

	var ok bool

	if value, found := instance.Get(name); !found {
		set = *rheltypes.NewSortedSet()
	} else if set, ok = value.(rheltypes.SortedSet); !ok {
		return nil, c.ErrWrap(fmt.Errorf("expected stream, got %T", value))
	}

	value = rheltypes.Integer(0)

	if !set.Add(key, score) {
		value = rheltypes.Integer(1)
	}

	instance.Set(name, set)

	return
}
