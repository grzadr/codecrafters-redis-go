package commands

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal"
	"github.com/codecrafters-io/redis-starter-go/rheltypes"
)

type stringFlag struct {
	set   bool
	value string
}

func (sf *stringFlag) Set(x string) error {
	sf.value = x
	sf.set = true

	return nil
}

func (sf stringFlag) String() string {
	return sf.value
}

type ConfigArgs struct {
	Dir        stringFlag
	DbFilename stringFlag
	Port       string
	ReplicaOf  stringFlag
}

func NewConfigArgs() (conf *ConfigArgs) {
	conf = &ConfigArgs{}
	flag.Var(&conf.Dir, "dir", "directory of db files")
	flag.Var(&conf.DbFilename, "dbfilename", "name of db file")
	flag.StringVar(&conf.Port, "port", "6379", "listen port number")
	flag.StringVar(&conf.Port, "p", "6379", "listen port number")
	flag.Var(&conf.ReplicaOf, "replicaof", "address of master")
	flag.Parse()

	return
}

func (conf *ConfigArgs) IsDirSet() bool {
	return conf.Dir.set
}

func (conf *ConfigArgs) IsDbFilenameSet() bool {
	return conf.DbFilename.set
}

func (conf *ConfigArgs) IsReplicaOf() bool {
	return conf.ReplicaOf.set
}

func (conf *ConfigArgs) RegisterArgs(config *rheltypes.SafeMap) {
	v := reflect.ValueOf(conf).Elem()
	t := reflect.TypeOf(conf).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		switch sf := field.Interface().(type) {
		case stringFlag:
			if !field.Field(0).Bool() {
				continue
			}

			config.SetString(strings.ToLower(fieldType.Name), sf.String(), 0)
		case string:
			config.SetString(strings.ToLower(fieldType.Name), sf, 0)
		default:
			continue
		}
	}
}

func (conf *ConfigArgs) Register(config *rheltypes.SafeMap) {
	conf.RegisterArgs(config)

	if !conf.ReplicaOf.set {
		config.SetString(
			"role",
			"master",
			0,
		)
		config.SetString(
			"master_replid",
			"8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb",
			0,
		)
		config.SetString(
			"master_repl_offset",
			"0",
			0,
		)
	} else {
		config.SetString(
			"role",
			"slave",
			0,
		)
	}
}

func (conf *ConfigArgs) DbFilePath() (dbPath string, err error) {
	if !conf.IsDbFilenameSet() {
		return
	}

	dbPath = conf.DbFilename.value

	if conf.IsDirSet() {
		err = os.MkdirAll(conf.Dir.value, defaultMkdirMode)
		dbPath = path.Join(conf.Dir.value, dbPath)
	}

	return
}

func readDbFile(dbPath string) (err error) {
	file, err := os.Open(dbPath)
	if err != nil {
		return fmt.Errorf("error reading content of %q: %w", dbPath, err)
	}

	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("error closing %q: %w", dbPath, closeErr)
		}
	}()

	iter := internal.NewByteIteratorFromFile(file)

	rdbFile, err := internal.ReadRdbFile(iter)

	data := GetDataMapInstance()

	for key, value := range rdbFile.Iter() {
		data.SetStringValue(
			key, value.Value, value.Expiry,
		)
	}

	return err
}

func closeWriter(
	file *os.File,
	writer *bufio.Writer,
	err *error,
) {
	if errFlush := writer.Flush(); errFlush != nil {
		errFlush = fmt.Errorf("error flushing writer: %w", errFlush)
		if *err == nil {
			*err = errFlush
		} else {
			*err = fmt.Errorf("%w; %w", *err, errFlush)
		}
	}

	if errClose := file.Close(); errClose != nil {
		errClose = fmt.Errorf("error closing file: %w", errClose)
		if *err == nil {
			*err = errClose
		} else {
			*err = fmt.Errorf("%w; %w", *err, errClose)
		}
	}
}

func createDbFile(dbPath string) (err error) {
	file, err := os.Create(dbPath)
	if err != nil {
		return
	}

	writer := bufio.NewWriter(file)

	err = internal.NewRdbfFile().WriteContent(writer)

	defer closeWriter(file, writer, &err)

	return
}

func (conf *ConfigArgs) InitDb() (err error) {
	dbPath, err := conf.DbFilePath()

	if err != nil || len(dbPath) == 0 {
		return
	}

	if _, err = os.Stat(dbPath); err == nil {
		err = readDbFile(dbPath)
	} else if os.IsNotExist(err) {
		err = createDbFile(dbPath)
	} else {
		return fmt.Errorf("error reading %q: %q", dbPath, err)
	}

	return
}

const defaultMkdirMode = 0o755

func Setup() (conf *ConfigArgs, err error) {
	conf = NewConfigArgs()

	conf.Register(GetConfigMapInstance())

	if err = conf.InitDb(); err != nil {
		err = fmt.Errorf("error during db init: %w", err)
	}

	return
}
