package main

import (
	"github.com/godyy/ggs/app/login/internal/app"
	_ "github.com/godyy/ggs/app/login/internal/handlers"
	"github.com/godyy/ggskit/utils"
)

func main() {
	app.Start()
	utils.ListenShutdown()
	app.Stop()
}
