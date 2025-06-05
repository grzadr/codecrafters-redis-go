package commands

import (
	"flag"
	"log"
	"os"
	"reflect"
	"strings"

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

const defaultMkdirMode = 0o755

func mkdir(dir string) {
	err := os.MkdirAll(dir, defaultMkdirMode)
	if err != nil {
		log.Fatalf("Error creating directory: %v\n", err)

		return
	}
}

func Setup() {
	conf := NewConfigArgs()

	conf.Register(GetConfigMapInstance())

	if dirFlag := conf.Dir; dirFlag.set {
		mkdir(dirFlag.String())
	}
}
