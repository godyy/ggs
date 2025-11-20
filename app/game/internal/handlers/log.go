package handlers

import (
	"github.com/godyy/ggs/internal/libs/logger"
	"github.com/godyy/glog"
)

var loggerInst glog.Logger

func init() {
	logger.RegisterAfterInitFunc(func(l glog.Logger) {
		loggerInst = l.Named("handlers")
	})
}
