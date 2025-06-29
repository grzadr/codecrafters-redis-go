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
	"DISCARD":  func() RhelCommand { return NewCmdDiscard() },
	"ECHO":     func() RhelCommand { return NewCmdEcho() },
	"EXEC":     func() RhelCommand { return NewCmdExec() },
	"GET":      func() RhelCommand { return NewCmdGet() },
	"INCR":     func() RhelCommand { return NewCmdIncr() },
	"INFO":     func() RhelCommand { return NewCmdInfo() },
	"KEYS":     func() RhelCommand { return NewCmdKeys() },
	"MULTI":    func() RhelCommand { return NewCmdMulti() },
	"PING":     func() RhelCommand { return NewCmdPing() },
	"PSYNC":    func() RhelCommand { return NewCmdPsync() },
	"REPLCONF": func() RhelCommand { return NewCmdReplconf() },
	"SET":      func() RhelCommand { return NewCmdSet() },
	"TYPE":     func() RhelCommand { return NewCmdType() },
	"WAIT":     func() RhelCommand { return NewCmdWait() },
	"XADD":     func() RhelCommand { return NewCmdXAdd() },
	"XRANGE":   func() RhelCommand { return NewCmdXRange() },
	"XREAD":    func() RhelCommand { return NewCmdXRead() },
}

func NewRhelCommand(name string) RhelCommand {
	if factory, exists := commandMap[strings.ToUpper(name)]; exists {
		return factory()
	}

	return BaseCommand(name)
}

type ParsedCommand struct {
	cmd   RhelCommand
	args  rheltypes.Array
	err   error
	size  int
	ack   int
	multi bool
	exec  bool
}

func newParsedCommandErr(err error) (parsed *ParsedCommand) {
	parsed = &ParsedCommand{err: fmt.Errorf("failed to parse command: %w", err)}

	return
}

func newParsedCommandFromArray(args rheltypes.Array) (parsed *ParsedCommand) {
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
	case CmdMulti:
		parsed.multi = true

	case CmdExec:
		parsed.exec = true
	}

	return
}

func newParsedCommand(raw rheltypes.RhelType) (parsed *ParsedCommand) {
	switch value := raw.(type) {
	case rheltypes.Array:
		parsed = newParsedCommandFromArray(value)

	case rheltypes.SimpleString, rheltypes.BulkString:

	default:
		parsed = newParsedCommandErr(
			fmt.Errorf("expected array, got %T", value),
		)
	}

	return
}

func newParsedCommandFromBytes(
	command []byte,
) iter.Seq[*ParsedCommand] {
	// log.Printf("command:\n%s", hex.Dump(command))
	return func(yield func(*ParsedCommand) bool) {
		tokens := rheltypes.NewTokenIterator(command)

		for {
			rawValue, err := rheltypes.RhelEncode(tokens)
			if err != nil {
				yield(
					newParsedCommandErr(fmt.Errorf("encoding error: %w", err)),
				)

				return
			}

			if tokens.IsDone() {
				return
			}

			parsed := newParsedCommand(rawValue)

			if parsed == nil {
				continue
			}

			if !yield(parsed) || parsed.err != nil {
				return
			}
		}
	}
}

func (p *ParsedCommand) Transaction(t *Transaction) {
	switch p.cmd.(type) {
	case CmdMulti:
		t = NewTransaction()

		return
	case CmdDiscard:

	case CmdExec:
		if tran != nil {
			results, parsed.args, err := tran.Exec()
			// TODO iterate results
		} else {
			parsed.args = nil
		}

		tran = nil
	}

	if t == nil {
		p.args = nil
	}

	tran = nil
}

func (p *ParsedCommand) Exec() (result *CommandResult) {
	result = &CommandResult{}

	cmd := p.cmd

	result.result, result.Err = cmd.Exec(p.args)
	if err := result.Err; err != nil {
		result.Err = fmt.Errorf(
			"failed to run command %s: %w",
			cmd.Name(),
			err,
		)

		return
	}

	result.Resend = cmd.Resend()
	result.ReplicaRespond = cmd.ReplicaRespond()
	result.Size = p.size
	result.Ack = p.ack

	return
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

func newCommandResultQueued() (result *CommandResult) {
	result.result = rheltypes.SimpleString("QUEUED")

	return
}

func (r CommandResult) Serialize() []byte {
	if r.result == nil {
		return nil
	}

	return r.result.Serialize()
}

const defaultTransactionCapacity = 16

type Transaction struct {
	cmds []*ParsedCommand
}

func NewTransaction() *Transaction {
	return &Transaction{
		cmds: make([]*ParsedCommand, 0, defaultTransactionCapacity),
	}
}

func (t Transaction) Exec() (results []CommandResult, responses rheltypes.Array, err error) {
	return
}

func ExecuteCommand(
	command []byte,
	tran *Transaction,
) iter.Seq[*CommandResult] {
	return func(yield func(*CommandResult) bool) {
		for parsed := range newParsedCommandFromBytes(command) {
			if err := parsed.err; err != nil {
				yield(&CommandResult{Err: NewCommandError(command, err)})

				return
			}

			var result *CommandResult

			if tran != nil {
				tran.cmds = append(tran.cmds, parsed)
				result = newCommandResultQueued()
			} else if result := newCommandResult(parsed); parsed.err != nil {
			}

			if err := parsed.err; err != nil {
				result.Err = NewCommandError(command, err)
			}

			if !yield(result) || result.Err != nil {
				return
			}

			switch p := parsed.cmd.(type) {
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
