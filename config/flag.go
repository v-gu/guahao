package config

import (
	"flag"
)

func parseFlags() {
	flag.StringVar(&configFile, "config", "settings.toml",
		"the config file to use")
	flag.Parse()
}
