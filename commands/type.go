package commands

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdType struct {
	BaseCommand
}

func NewCmdType() CmdType {
	return CmdType{BaseCommand: BaseCommand("TYPE")}
}

func (c CmdType) Exec(
	args rheltypes.Array,
) (valueType rheltypes.RhelType, err error) {
	key := args.At(0)
	if key == nil {
		return nil, c.ErrWrap(fmt.Errorf("missing key"))
	}

	instance := GetDataMapInstance()

	valueType = rheltypes.SimpleString("none")

	if value, found := instance.Get(key.String()); found {
		valueType = rheltypes.SimpleString(value.TypeName())
	}

	return
}
