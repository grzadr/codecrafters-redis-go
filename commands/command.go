package commands

import (
	"encoding/hex"
	"fmt"
	"iter"
	"strings"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/pubsub"
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

type RhelCommand interface {
	isRhelCommand()
	Name() string
	ErrWrap(input error) error
	Exec(args rheltypes.Array) (rheltypes.RhelType, error)
	Resend() bool
	ReplicaRespond() bool
	AllowedInSubscription() bool
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

func (c BaseCommand) AllowedInSubscription() bool { return false }

func (BaseCommand) isRhelCommand() {}

var commandMap = map[string]func() RhelCommand{
	"BLPOP":       func() RhelCommand { return NewCmdBLPop() },
	"CONFIG":      func() RhelCommand { return NewCmdConfig() },
	"DISCARD":     func() RhelCommand { return NewCmdDiscard() },
	"ECHO":        func() RhelCommand { return NewCmdEcho() },
	"EXEC":        func() RhelCommand { return NewCmdExec() },
	"GET":         func() RhelCommand { return NewCmdGet() },
	"INCR":        func() RhelCommand { return NewCmdIncr() },
	"INFO":        func() RhelCommand { return NewCmdInfo() },
	"KEYS":        func() RhelCommand { return NewCmdKeys() },
	"LLEN":        func() RhelCommand { return NewCmdLLen() },
	"LPOP":        func() RhelCommand { return NewCmdLPop() },
	"LPUSH":       func() RhelCommand { return NewCmdLPush() },
	"LRANGE":      func() RhelCommand { return NewCmdLRange() },
	"MULTI":       func() RhelCommand { return NewCmdMulti() },
	"PING":        func() RhelCommand { return NewCmdPing() },
	"PSYNC":       func() RhelCommand { return NewCmdPsync() },
	"PUBLISH":     func() RhelCommand { return NewCmdPublish() },
	"REPLCONF":    func() RhelCommand { return NewCmdReplconf() },
	"RPUSH":       func() RhelCommand { return NewCmdRPush() },
	"SET":         func() RhelCommand { return NewCmdSet() },
	"SUBSCRIBE":   func() RhelCommand { return NewCmdSubscribe() },
	"TYPE":        func() RhelCommand { return NewCmdType() },
	"UNSUBSCRIBE": func() RhelCommand { return NewCmdUnsubscribe() },
	"WAIT":        func() RhelCommand { return NewCmdWait() },
	"XADD":        func() RhelCommand { return NewCmdXAdd() },
	"XRANGE":      func() RhelCommand { return NewCmdXRange() },
	"XREAD":       func() RhelCommand { return NewCmdXRead() },
	"ZADD":        func() RhelCommand { return NewCmdZAdd() },
	"ZCARD":       func() RhelCommand { return NewCmdZCard() },
	"ZRANGE":      func() RhelCommand { return NewCmdZRange() },
	"ZRANK":       func() RhelCommand { return NewCmdZRank() },
	"ZREM":        func() RhelCommand { return NewCmdZRem() },
	"ZSCORE":      func() RhelCommand { return NewCmdZScore() },
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
	sub   bool
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

func (p *ParsedCommand) Commit(t **Transaction) (err error) {
	switch p.cmd.(type) {
	case CmdMulti:
		*t = NewTransaction()
	case CmdSubscribe:
		if *t == nil {
			*t = NewTransaction()
		}

		p.sub = true
	case CmdUnsubscribe:
		p.args.Append(rheltypes.Integer((*t).subscriptions[p.args.First().String()]))
	case CmdDiscard:
		if *t == nil {
			p.args = nil
		}

		*t = nil
	case CmdExec:
		if *t != nil {
			_, p.args, err = (*t).Exec()
		} else {
			p.args = nil
		}

		*t = nil
	}

	p.sub = p.sub || (*t).IsSubscribed()

	return err
}

func (p *ParsedCommand) Exec(t **Transaction) (result *CommandResult) {
	result = &CommandResult{}
	if !p.verifySubscription(t) {
		result.result = p.newErrorForbiddenCommand()

		return result
	}

	cmd := p.cmd

	result.result, result.Err = cmd.Exec(p.args)
	if err := result.Err; err != nil {
		result.Err = fmt.Errorf(
			"failed to run command %s: %w",
			cmd.Name(),
			err,
		)

		return result
	}

	if *t != nil {
		(*t).digestSubscription(cmd, result)
		result.Sub = (*t).SubStart
	}

	result.Resend = cmd.Resend()
	result.ReplicaRespond = cmd.ReplicaRespond()
	result.Size = p.size
	result.Ack = p.ack

	return result
}

func (p *ParsedCommand) verifySubscription(t **Transaction) bool {
	return p.cmd.AllowedInSubscription() || *t == nil || !(*t).IsSubscribed()
}

func (p *ParsedCommand) newErrorForbiddenCommand() rheltypes.Error {
	return rheltypes.NewGenericError(
		fmt.Errorf(
			"Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context",
			strings.ToLower(p.cmd.Name()),
		),
	)
}

type CommandResult struct {
	result         rheltypes.RhelType
	KeepConnection bool
	Resend         bool
	ReplicaRespond bool
	Err            error
	Size           int
	Ack            int
	Sub            bool
}

func newCommandResultQueued() (result *CommandResult) {
	result = &CommandResult{result: rheltypes.SimpleString("QUEUED")}

	return
}

func NewCommandErrorResponse(content []byte, message error) *CommandResult {
	return &CommandResult{Err: CommandError{content: content, message: message}}
}

func (r CommandResult) Serialize() []byte {
	if r.result == nil {
		return nil
	}

	return r.result.Serialize()
}

const defaultTransactionCapacity = 16

type Transaction struct {
	cmds          []*ParsedCommand
	subscriptions map[string]int
	lock          sync.RWMutex
	SubStart      bool
}

func NewTransaction() *Transaction {
	return &Transaction{
		cmds:          make([]*ParsedCommand, 0, defaultTransactionCapacity),
		subscriptions: make(map[string]int, defaultTransactionCapacity),
	}
}

func (t *Transaction) Exec() (results []*CommandResult, responses rheltypes.Array, err error) {
	results = make([]*CommandResult, len(t.cmds))
	responses = make(rheltypes.Array, len(t.cmds))

	for i, c := range t.cmds {
		r := c.Exec(&t)

		if r.Err != nil {
			err = r.Err

			return
		}

		results[i] = r
		responses[i] = r.result
	}

	return
}

func (t *Transaction) IsSubscribed() bool {
	if t == nil {
		return false
	}

	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.numSubscriptions() > 0
}

func (t *Transaction) IterSubscriptions() iter.Seq2[string, *pubsub.Subscription] {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return func(yield func(string, *pubsub.Subscription) bool) {
		st := pubsub.GetStreamManager()

		for name, id := range t.subscriptions {
			if !yield(name, st.GetSubscription(name, id)) {
				return
			}
		}
	}
}

func (t *Transaction) numSubscriptions() int {
	return len(t.subscriptions)
}

func (t *Transaction) digestSubscription(
	cmd RhelCommand,
	result *CommandResult,
) {
	t.lock.Lock()
	defer t.lock.Unlock()

	switch cmd.(type) {
	case CmdSubscribe:
		arr := result.result.(rheltypes.Array)
		key := arr.At(1).String()
		id, _ := arr.At(cmdSubscribeResultNumPos).Integer()

		(*t).subscriptions[key] = id

		(*t).SubStart = (*t).numSubscriptions() == 1

		arr.Set(cmdSubscribeResultNumPos, rheltypes.Integer(t.numSubscriptions()))

		result.result = arr
	case CmdUnsubscribe:
		arr := result.result.(rheltypes.Array)
		key := arr.At(1).String()

		delete((*t).subscriptions, key)

		arr.Set(cmdSubscribeResultNumPos, rheltypes.Integer(t.numSubscriptions()))
	case CmdPing:
		if t.numSubscriptions() == 0 {
			return
		}

		result.result = rheltypes.NewArrayFromStrings([]string{"pong", ""})
	}
}

func ExecuteCommand(
	command []byte,
	tran **Transaction,
) iter.Seq[*CommandResult] {
	return func(yield func(*CommandResult) bool) {
		for parsed := range newParsedCommandFromBytes(command) {
			if err := parsed.err; err != nil {
				yield(NewCommandErrorResponse(command, err))

				return
			}

			if err := parsed.Commit(tran); err != nil {
				yield(NewCommandErrorResponse(command, err))

				return
			}

			var result *CommandResult

			if *tran != nil && !parsed.multi && !parsed.sub {
				(*tran).cmds = append((*tran).cmds, parsed)
				result = newCommandResultQueued()
			} else if result = parsed.Exec(tran); result.Err != nil {
				result = NewCommandErrorResponse(command, result.Err)
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
