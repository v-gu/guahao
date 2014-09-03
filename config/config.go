package config

import (
	"errors"
	"fmt"
	"io/ioutil"

	toml "github.com/BurntSushi/toml"

	log "github.com/v-gu/guahao/log"
)

var (
	configFile string
	tomlString string
)

var All Config = Config{StorePath: "run", NamedLogger: log.NamedLogger{"config"}}

type Config struct {
	StorePath string
	Disabled  []string

	metaData     toml.MetaData
	sectionPrims map[string]toml.Primitive

	log.NamedLogger
}

func init() {
	parseFlags()
	getGlobalConfig()
}

//
func getGlobalConfig() {
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(err)
	}
	tomlString = string(bytes)

	// read global config entries
	_, err = toml.Decode(tomlString, &All)
	if err != nil {
		panic(err)
	}
	All.Debugf(log.DEBUG_CONFIG, "globalconf -> %#v\n", All)

	// read section config entries
	All.metaData, err = toml.Decode(tomlString, &All.sectionPrims)
	if err != nil {
		panic(err)
	}
}

//
func (c *Config) UnmarshalConfig(section string, config interface{}) (err error) {
	for _, k := range c.Disabled {
		if k == section {
			err = errors.New(
				fmt.Sprintf("section '%v' was disabled by config file", section))
			return
		}
	}
	err = c.metaData.PrimitiveDecode(c.sectionPrims[section], config)
	c.Debugf(log.DEBUG_CONFIG, "config: decoded: %v\n", config)
	return
}
