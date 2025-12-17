package main

import (
	"github.com/godyy/ggs/app/client/internal/conf"
	"github.com/godyy/ggs/app/client/internal/mode"
	_ "github.com/godyy/ggs/app/client/internal/mode/client"
	_ "github.com/godyy/ggs/app/client/internal/mode/robot"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/utils"
	"github.com/godyy/glog"
	"github.com/godyy/gutils/flags"
)

func main() {
	// 解析 flags
	flags.Parse()

	// 初始化 logger
	logger.Init(&logger.Config{
		Level:       glog.DebugLevel,
		Caller:      true,
		Development: true,
		EnableStd:   true,
	})

	mode := mode.CreateMode(conf.Mode)
	mode.Start()
	flags.Reset()
	utils.ListenShutdown()
	mode.Stop()
}
