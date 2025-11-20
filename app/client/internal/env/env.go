package env

import (
	"log"

	"github.com/godyy/ggs/internal/env"
	"github.com/godyy/ggs/internal/libs/flags"
)

var (
	LoginURLRoot string // login url root
	AgentAddr    string // server address
	SignKeyPath  string // auth private key path
	Mode         string // mode: client or robot

)

func Init() {
	env.Init()

	LoginURLRoot, _ = flags.GetValue[string]("login-url-root")
	if LoginURLRoot == "" {
		log.Fatalf("-login-url-root is empty")
	}
	AgentAddr, _ = flags.GetValue[string]("agent-addr")
	if AgentAddr == "" {
		log.Fatalf("-agent-addr is empty")
	}
	Mode, _ = flags.GetValue[string]("mode")
	if Mode == "" {
		log.Fatalf("-mode is empty")
	}
	SignKeyPath, _ = flags.GetValue[string]("sign-key-path")
	if SignKeyPath == "" {
		log.Fatalf("-sign-key-path is empty")
	}
}

func init() {
	flags.String("login-url-root", "", "login url root")
	flags.String("agent-addr", "", "agent address")
	flags.String("sign-key-path", "", "sign key path")
	flags.String("mode", "client", "client or robot")
}
