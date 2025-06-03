package commands

import (
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type CommandError struct {
	content []byte
	message error
}

func (e CommandError) Error() string {
	return fmt.Sprintf("failed to process command %q: %s\n%s",
		string(e.content), e.message, hex.Dump(e.content))
}

func NewCommandError(content []byte, message error) error {
	return CommandError{content: content, message: message}
}

type RhelCommand interface {
	isRhelCommand()
	Name() string
	Exec(rheltypes.Array) (rheltypes.RhelType, error)
}

type UnknownCommand string

func (UnknownCommand) isRhelCommand() {}

func (c UnknownCommand) Name() string {
	return string(c)
}

func (c UnknownCommand) Exec(
	value rheltypes.Array,
) (rheltypes.RhelType, error) {
	return nil, fmt.Errorf("command %q not found", c.Name())
}

func NewRhelCommand(
	name string,
) (cmd RhelCommand) {
	switch strings.ToUpper(name) {
	case "PING":
		return CmdPing{}
	case "ECHO":
		return CmdEcho{}
	default:
		return UnknownCommand(name)
	}
}

func parseCommand(
	command []byte,
) (cmd RhelCommand, args rheltypes.Array, err error) {
	wrap := func(newErr error) error {
		if newErr != nil {
			return fmt.Errorf("error in parseCommand: %w", newErr)
		}
		return nil
	}
	tokens, err := rheltypes.NewTokenIterator(command)
	if err != nil {
		return nil, nil, wrap(fmt.Errorf("tokenization error: %w", err))
	}

	rawValue, err := rheltypes.RhelEncode(tokens)
	if err != nil {
		return nil, nil, wrap(fmt.Errorf("encoding error: %w", err))
	}

	switch value := rawValue.(type) {
	case rheltypes.Array:
		return NewRhelCommand(value[0].String()), value[1:], nil
	default:
		return nil, nil, wrap(fmt.Errorf("expected array, got %T", value))
	}
}

func ExecuteCommand(command []byte) (result rheltypes.RhelType, err error) {
	log.Printf("Command %q:\n%s", command, hex.Dump(command))

	cmd, args, err := parseCommand(command)
	if err != nil {
		return nil, NewCommandError(command, err)
	}

	result, err = cmd.Exec(args)
	if err != nil {
		return nil, NewCommandError(command, err)
	}
	return
}
