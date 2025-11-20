package persist

import (
	liblogger "github.com/godyy/ggs/internal/libs/logger"
	"github.com/godyy/glog"
)

var (
	logger glog.Logger
)

func init() {
	liblogger.RegisterAfterInitFunc(func(l liblogger.Logger) {
		logger = l.Named("actor.persist")
	})
}
