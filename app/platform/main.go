package main

import (
	"github.com/godyy/ggs/app/platform/internal/app"
	"github.com/godyy/ggs/internal/utils"

	_ "github.com/godyy/ggs/app/platform/internal/data/repository"
	_ "github.com/godyy/ggs/app/platform/internal/handlers"
)

func main() {
	app.Start()
	utils.ListenShutdown()
	app.Stop()
}
