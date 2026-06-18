package handler

import (
	"github.com/godyy/ggs/internal/base/logger"
)

var loggerInst logger.Logger

func init() {
	logger.RegisterAfterInitFunc(func(l logger.Logger) {
		loggerInst = l.Named("handlers")
	})
}
