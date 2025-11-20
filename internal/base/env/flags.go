package env

import (
	"github.com/godyy/ggs/internal/libs/flags"
)

func FlagName(name string) string {
	return "env-" + name
}

func (env *envImpl) applyFlags() {
	env.stage, _ = flags.GetValue[string](FlagName("stage"))
	env.debug, _ = flags.GetValue[bool](FlagName("debug"))
}

func init() {
	flags.String(FlagName("stage"), StageDev, "stage")
	flags.Bool(FlagName("debug"), false, "debug")
	flags.AddParsedFunc(env.applyFlags)
}
