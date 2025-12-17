package main

import (
	"github.com/godyy/ggs/app/platform/internal/app"
	_ "github.com/godyy/ggs/app/platform/internal/handlers"
	"github.com/godyy/ggs/internal/utils"
)

func main() {
	app.Start()
	utils.ListenShutdown()
	app.Stop()
}
