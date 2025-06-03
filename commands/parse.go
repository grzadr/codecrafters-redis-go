package commands

import (
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type ContentError struct {
	content []byte
	message error
}

func (e ContentError) Error() string {
	return fmt.Sprintf("malformed content %q: %s\n%s",
		string(e.content), e.message, hex.Dump(e.content))
}

func NewContentError(content []byte, message error) error {
	return ContentError{content: content, message: message}
}

type RhelCommand interface {
	isRhelCommand()
	Name() string
	Exec(rheltypes.RhelType) (rheltypes.RhelType, error)
}

type UnknownCommand string

func (UnknownCommand) isRhelCommand() {}

func (c UnknownCommand) Name() string {
	return string(c)
}

func (c UnknownCommand) Exec(
	value rheltypes.RhelType,
) (rheltypes.RhelType, error) {
	return nil, fmt.Errorf("command not found %s", c.Name())
}

var commands = map[string]RhelCommand{
	"PING": CmdPing{},
	"ECHO": CmdEcho{},
}

func NewRhelCommand(
	value rheltypes.RhelType,
) (cmd RhelCommand) {
	cmdStr := strings.ToUpper(value.First().String())
	var found bool

	if cmd, found = commands[cmdStr]; !found {
		return UnknownCommand(cmdStr)
	}

	return
}

func ExecuteCommand(content []byte) (result rheltypes.RhelType, err error) {
	log.Printf("Command %q:\n%s", content, hex.Dump(content))
	tokens, err := rheltypes.NewTokenIterator(content)
	if err != nil {
		return nil, fmt.Errorf(
			"tokenization error: %w",
			NewContentError(content, err),
		)
	}

	value, err := rheltypes.RhelEncode(tokens)
	if err != nil {
		return nil, NewContentError(
			content,
			fmt.Errorf("encoding error: %w", err),
		)
	}
	cmd := NewRhelCommand(value)
	result, err = cmd.Exec(value)
	if err != nil {
		return nil, fmt.Errorf(
			"command %s execution error: %w",
			cmd.Name(),
			NewContentError(content, err),
		)
	}

	return
}
