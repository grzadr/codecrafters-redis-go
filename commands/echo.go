package commands

import (
	"fmt"
	"log"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdEcho struct{}

func (CmdEcho) isRhelCommand() {}

func (CmdEcho) Name() string {
	return "ECHO"
}

func (e CmdEcho) Exec(
	args rheltypes.RhelType,
) (value rheltypes.RhelType, err error) {
	log.Println(e.Name(), args.String())
	var ok bool
	switch v := args.(type) {
	case rheltypes.Array:
		if value, ok = v.At(1); !ok {
			err = fmt.Errorf("expected 2 values, got %d", len(v))
		}
	default:
		err = fmt.Errorf("expected %T, got %T", rheltypes.Array{}, v)
	}
	return
}
