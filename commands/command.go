package commands

import (
	"encoding/hex"
	"fmt"
	"iter"
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
	Resend() bool
	ReplicaRespond() bool
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

func (c BaseCommand) Resend() bool { return false }

func (c BaseCommand) ReplicaRespond() bool { return false }

func (BaseCommand) isRhelCommand() {}

var commandMap = map[string]func() RhelCommand{
	"CONFIG":   func() RhelCommand { return NewCmdConfig() },
	"ECHO":     func() RhelCommand { return NewCmdEcho() },
	"GET":      func() RhelCommand { return NewCmdGet() },
	"INFO":     func() RhelCommand { return NewCmdInfo() },
	"KEYS":     func() RhelCommand { return NewCmdKeys() },
	"PING":     func() RhelCommand { return NewCmdPing() },
	"PSYNC":    func() RhelCommand { return NewCmdPsync() },
	"REPLCONF": func() RhelCommand { return NewCmdReplconf() },
	"SET":      func() RhelCommand { return NewCmdSet() },
	"TYPE":     func() RhelCommand { return NewCmdType() },
	"WAIT":     func() RhelCommand { return NewCmdWait() },
	"XADD":     func() RhelCommand { return NewCmdXAdd() },
	"XRANGE":   func() RhelCommand { return NewCmdXRange() },
	"XREAD":    func() RhelCommand { return NewCmdXRead() },
	"INCR":     func() RhelCommand { return NewCmdIncr() },
	"MULTI":    func() RhelCommand { return NewCmdMulti() },
	"EXEC":     func() RhelCommand { return NewCmdExec() },
}

func NewRhelCommand(name string) RhelCommand {
	if factory, exists := commandMap[strings.ToUpper(name)]; exists {
		return factory()
	}

	return BaseCommand(name)
}

type ParsedCommand struct {
	cmd  RhelCommand
	args rheltypes.Array
	err  error
	size int
	ack  int
}

func NewParsedCommandErr(err error) (parsed *ParsedCommand) {
	parsed = &ParsedCommand{err: fmt.Errorf("failed to parse command: %w", err)}

	return
}

func NewParsedCommandFromArray(args rheltypes.Array) (parsed *ParsedCommand) {
	parsed = &ParsedCommand{
		cmd:  NewRhelCommand(args[0].String()),
		args: args[1:],
		size: args.Size(),
	}

	switch parsed.cmd.(type) {
	case CmdReplconf:
		if parsed.args.Cmd() == "ACK" {
			parsed.ack, parsed.err = parsed.args.At(1).Integer()
		}
		// case CmdMulti:
	}

	return
}

func NewParsedCommand(raw rheltypes.RhelType) (parsed *ParsedCommand) {
	switch value := raw.(type) {
	case rheltypes.Array:
		parsed = NewParsedCommandFromArray(value)

	case rheltypes.SimpleString, rheltypes.BulkString:

	default:
		parsed = NewParsedCommandErr(
			fmt.Errorf("expected array, got %T", value),
		)
	}

	return
}

func parseCommand(
	command []byte,
) iter.Seq[*ParsedCommand] {
	// log.Printf("command:\n%s", hex.Dump(command))
	return func(yield func(*ParsedCommand) bool) {
		tokens := rheltypes.NewTokenIterator(command)

		// offset := tokens.Offset()

		for {
			rawValue, err := rheltypes.RhelEncode(tokens)
			if err != nil {
				yield(
					NewParsedCommandErr(fmt.Errorf("encoding error: %w", err)),
				)

				return
			}

			// size := tokens.Offset() - offset
			// = tokens.Offset()

			if tokens.IsDone() {
				return
			}

			parsed := NewParsedCommand(rawValue)

			if parsed == nil {
				continue
			}

			if !yield(parsed) || parsed.err != nil {
				return
			}
		}
	}
}

type CommandResult struct {
	result         rheltypes.RhelType
	KeepConnection bool
	Resend         bool
	ReplicaRespond bool
	Err            error
	Size           int
	Ack            int
}

func (r CommandResult) Serialize() []byte {
	if r.result == nil {
		return nil
	}

	return r.result.Serialize()
}

func ExecuteCommand(command []byte) iter.Seq[*CommandResult] {
	return func(yield func(*CommandResult) bool) {
		for parsed := range parseCommand(command) {
			result := &CommandResult{}
			if err := parsed.err; err != nil {
				result.Err = NewCommandError(command, err)
				yield(result)

				return
			}

			cmd := parsed.cmd

			result.result, result.Err = cmd.Exec(parsed.args)
			if err := result.Err; err != nil {
				result.Err = NewCommandError(command, err)
				yield(result)

				return
			}

			result.Resend = cmd.Resend()
			result.ReplicaRespond = cmd.ReplicaRespond()
			result.Size = parsed.size
			result.Ack = parsed.ack

			if !yield(result) {
				return
			}

			switch p := cmd.(type) {
			case CmdPsync:
				if !yield(
					&CommandResult{result: p.RenderFile(), KeepConnection: true},
				) {
					return
				}
			}
		}
	}
}

// func ExecuteCommand(command []byte) (result rheltypes.RhelType, err error) {
// 	cmd, args, err := parseCommand(command)
// 	if err != nil {
// 		return nil, NewCommandError(command, err)
// 	}

// 	result, err = cmd.Exec(args)
// 	if err != nil {
// 		return nil, NewCommandError(command, err)
// 	}

// 	if cmd.(CmdPsync)

// 	return
// }
