package conf

import (
	"log"

	"github.com/godyy/ggskit/base/flags"
)

func applyFlags() {
	LoginURLRoot, _ = flags.GetValue[string]("login-url-root")
	if LoginURLRoot == "" {
		log.Fatal("-login-url-root is empty")
	}
	AgentAddr, _ = flags.GetValue[string]("agent-addr")
	if AgentAddr == "" {
		log.Fatal("-agent-addr is empty")
	}
	Mode, _ = flags.GetValue[string]("mode")
	if Mode == "" {
		log.Fatal("-mode is empty")
	}
	SignKeyPath, _ = flags.GetValue[string]("sign-key-path")
	if SignKeyPath == "" {
		log.Fatal("-sign-key-path is empty")
	}
}

func init() {
	flags.String("login-url-root", "", "login url root")
	flags.String("agent-addr", "", "agent address")
	flags.String("sign-key-path", "", "sign key path")
	flags.String("mode", "client", "client or robot")
	flags.AddParsedFunc(applyFlags)
}
