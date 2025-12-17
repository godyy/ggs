package main

import (
	"github.com/godyy/ggs/app/agent/internal/app"
	_ "github.com/godyy/ggs/app/agent/internal/infra/agent"
	"github.com/godyy/ggs/internal/utils"
)

func main() {
	app.Start()
	utils.ListenShutdown()
	app.Stop()
}
