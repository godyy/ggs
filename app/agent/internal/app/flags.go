package app

import "github.com/godyy/ggs/internal/libs/flags"

func init() {
	flags.String("config-path", "./configs/dev.toml", "config path")
}

func configPath() string {
	path, _ := flags.GetValue[string]("config-path")
	return path
}
