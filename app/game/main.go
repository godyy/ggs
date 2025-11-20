package main

import (
	"github.com/godyy/ggs/app/game/internal/app"
	"github.com/godyy/ggs/internal/utils"

	_ "github.com/godyy/ggs/app/game/internal"
)

func main() {
	app.Start()
	utils.ListenShutdown()
	app.Stop()
}
