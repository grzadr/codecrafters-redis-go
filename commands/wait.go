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
	required, _ := args.At(0).Integer()
	timeout, _ := args.At(1).Integer()

	log.Printf("WAIT for %d or %d mili", required, timeout)

	if required == 0 {
		return rheltypes.Integer(0), nil
	}

	conn := connection.GetConnectionPool()
	result := make(chan int, 1)

	go func() {
		defer close(result)

		after := time.After(time.Duration(timeout) * time.Millisecond)

		ticker := time.NewTicker(defaultWaitTicker)
		defer ticker.Stop()

		for {
			select {
			case <-after:
				log.Println("WAIT timeout")
				result <- conn.NumAcknowledged()

				return
			case <-ticker.C:
				if ack := conn.NumAcknowledged(); ack >= required {
					log.Printf("ack passed: %d", ack)
					result <- ack

					return
				}
			}
		}
	}()

	acknowledged := <-result

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
