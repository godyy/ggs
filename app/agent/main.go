package main

import (
	_ "github.com/godyy/ggs/app/agent/internal/agent"
	"github.com/godyy/ggs/app/agent/internal/app"
	"github.com/godyy/ggs/internal/utils"
)

func main() {
	app.Start()
	utils.ListenShutdown()
	app.Stop()
}
