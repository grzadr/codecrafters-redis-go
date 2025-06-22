package commands

import (
	"log"
	"time"

	"github.com/codecrafters-io/redis-starter-go/connection"
	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

var defaultWaitTicker = 10 * time.Millisecond

type CmdWait struct {
	BaseCommand
}

func NewCmdWait() CmdWait {
	return CmdWait{BaseCommand: BaseCommand("WAIT")}
}

func (c CmdWait) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	log.Println(args.String())
	required, _ := args.At(1).Integer()
	timeout, _ := args.At(1).Integer()

	conn := connection.GetConnectionPool()
	c1 := make(chan int, 1)

	go func() {
		defer close(c1)

		after := time.After(time.Duration(timeout) * time.Millisecond)

		ticker := time.NewTicker(defaultWaitTicker)
		defer ticker.Stop()

		for {
			select {
			case <-after:
				c1 <- conn.NumAcknowledged()

				return
			case <-ticker.C:
				acknowledged := conn.NumAcknowledged()
				if acknowledged >= required {
					c1 <- acknowledged

					return
				}
			}
		}
	}()

	acknowledged := <-c1

	log.Printf("acknowledged %d\n", acknowledged)

	return rheltypes.Integer(
		acknowledged,
	), nil
}

func (c CmdWait) Render(id, offset string) (cmd rheltypes.Array) {
	return rheltypes.NewArrayFromStrings(
		[]string{string(c.BaseCommand), id, offset},
	)
}
