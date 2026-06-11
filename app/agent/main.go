package main

import (
	"github.com/godyy/ggs/app/agent/internal/app"
	_ "github.com/godyy/ggs/app/agent/internal/infra/agent"
	"github.com/godyy/ggskit/utils"
)

func main() {
	app.Start()
	utils.ListenShutdown()
	app.Stop()
}
