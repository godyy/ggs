package main

import (
	"github.com/godyy/ggs/app/login/internal/app"
	"github.com/godyy/ggs/internal/utils"

	_ "github.com/godyy/ggs/app/login/internal/data/repository"
	_ "github.com/godyy/ggs/app/login/internal/handlers"
)

func main() {
	app.Start()
	utils.ListenShutdown()
	app.Stop()
}
