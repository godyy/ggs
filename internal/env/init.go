package env

import (
	"github.com/godyy/ggs/internal/libs/flags"
)

func Init() {
	stage, _ = flags.GetValue[string]("env-stage")
	debug, _ = flags.GetValue[bool]("env-debug")
}

func init() {
	flags.String("env-stage", StageDev, "stage")
	flags.Bool("env-debug", false, "debug")
}
