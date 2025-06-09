package commands

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
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
}

func NewConfigArgs() (conf *ConfigArgs) {
	conf = &ConfigArgs{}
	flag.Var(&conf.Dir, "dir", "directory of db files")
	flag.Var(&conf.DbFilename, "dbfilename", "name of db file")
	flag.Parse()

	return
}

func (conf *ConfigArgs) IsDirSet() bool {
	return conf.Dir.set
}

func (conf *ConfigArgs) IsDbFilenameSet() bool {
	return conf.DbFilename.set
}

func (conf *ConfigArgs) Register(config *rheltypes.SafeMap) {
	v := reflect.ValueOf(conf).Elem()
	t := reflect.TypeOf(conf).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.Field(0).Bool() {
			continue
		}

		switch sf := field.Interface().(type) {
		case stringFlag:
			config.SetString(strings.ToLower(fieldType.Name), sf.String(), 0)
		}
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

func dump(name string) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return fmt.Errorf("failed to dump %q: %w", name, err)
	}

	log.Printf("%d bytes,\n%s", len(data), hex.Dump(data))

	return nil
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

	if err = dump(dbPath); err != nil {
		return err
	}

	rdbFile, err := internal.ReadRdbFile(iter)

	data := GetDataMapInstance()

	for key, value := range rdbFile.Iter() {
		data.SetStringValue(
			key, value.Value, value.Expiry,
		)
	}

	log.Printf("%v\n", rdbFile)

	return err
}

func createDbFile(dbPath string) {
}

func (conf *ConfigArgs) InitDb() (err error) {
	dbPath, err := conf.DbFilePath()

	if err != nil || len(dbPath) == 0 {
		return
	}

	if _, err = os.Stat(dbPath); err == nil {
		err = readDbFile(dbPath)
	} else if os.IsNotExist(err) {
		createDbFile(dbPath)
	} else {
		return fmt.Errorf("error reading %q: %q", dbPath, err)
	}

	return
}

const defaultMkdirMode = 0o755

func Setup() (err error) {
	conf := NewConfigArgs()

	conf.Register(GetConfigMapInstance())

	if err = conf.InitDb(); err != nil {
		err = fmt.Errorf("error during db init: %w", err)
	}

	return
}
