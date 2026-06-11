package app

import (
	"github.com/godyy/ggskit/base/flags"
)

func init() {
	flags.String("config-path", "./configs/dev.toml", "config path")
}

func configPath() string {
	path, _ := flags.GetValue[string]("config-path")
	return path
}
