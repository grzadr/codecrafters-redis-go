package commands

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

const (
	defaultMapCapacity        = 1024
	defaultMapCleanupInterval = 1 * time.Minute
)

var (
	dataMap    *rheltypes.SafeMap
	configMap  *rheltypes.SafeMap
	dataOnce   sync.Once
	configOnce sync.Once
)

func GetDataMapInstance() *rheltypes.SafeMap {
	dataOnce.Do(func() {
		dataMap = rheltypes.NewSafeMap(defaultMapCleanupInterval)
	})

	return dataMap
}

func GetConfigMapInstance() *rheltypes.SafeMap {
	configOnce.Do(func() {
		configMap = rheltypes.NewSafeMap(0)
	})

	return configMap
}

func CloseMaps() {
	GetConfigMapInstance().Close()
	GetDataMapInstance().Close()
}

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
	ErrWrap(input error) error
	Exec(args rheltypes.Array) (rheltypes.RhelType, error)
}

type BaseCommand string

func (c BaseCommand) Name() string {
	return string(c)
}

func (c BaseCommand) ErrWrap(input error) (err error) {
	if input != nil {
		err = fmt.Errorf("failed to run %q command: %w", c.Name(), input)
	}

	return
}

func (c BaseCommand) Exec(
	value rheltypes.Array,
) (rheltypes.RhelType, error) {
	return nil, c.ErrWrap(fmt.Errorf("command %q not found", c.Name()))
}

func (BaseCommand) isRhelCommand() {}

var commandMap = map[string]func() RhelCommand{
	"PING":   func() RhelCommand { return NewCmdPing() },
	"ECHO":   func() RhelCommand { return NewCmdEcho() },
	"SET":    func() RhelCommand { return NewCmdSet() },
	"GET":    func() RhelCommand { return NewCmdGet() },
	"CONFIG": func() RhelCommand { return NewCmdConfig() },
	"KEYS":   func() RhelCommand { return NewCmdKeys() },
}

func NewRhelCommand(name string) RhelCommand {
	if factory, exists := commandMap[strings.ToUpper(name)]; exists {
		return factory()
	}

	return BaseCommand(name)
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
