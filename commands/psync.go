package commands

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/internal"
	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CmdPsync struct {
	BaseCommand
}

func NewCmdPsync() CmdPsync {
	return CmdPsync{BaseCommand: BaseCommand("PSYNC")}
}

func (c CmdPsync) Exec(
	args rheltypes.Array,
) (value rheltypes.RhelType, err error) {
	config := GetConfigMapInstance()

	id, _ := config.Get("master_replid")
	offset, _ := config.Get("master_repl_offset")

	return rheltypes.SimpleString(
		fmt.Sprintf("+FULLRESYNC %s %s", id, offset),
	), nil
}

func (c CmdPsync) Render(id, offset string) (cmd rheltypes.Array) {
	return rheltypes.NewArrayFromStrings(
		[]string{string(c.BaseCommand), id, offset},
	)
}

func (c CmdPsync) RenderFile() (content rheltypes.RhelType) {
	var buffer bytes.Buffer
	buf := bufio.NewWriter(&buffer)

	internal.NewRdbfFile().WriteContent(buf)
	buf.Flush()

	return rheltypes.NewBulkStringFromBytes(buffer.Bytes())
}
